package dishRepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"
	"itsTasty/pkg/api/domain"
	"log"
	"time"
)

const codeDuplicateEntry = 1062

var alreadyExists = errors.New("already exists")

//TODO: double check that all internal functions wrap the error

// sqlContextPreparer is an interface provided both by transaction and standard db connection
type sqlContextPreparer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type basicDishInfoDTO struct {
	Name     string
	ServedAt string
}

type MysqlRepo struct {
	db *sql.DB
}

func NewMysqlRepo(db *sql.DB) (*MysqlRepo, error) {

	//create db tables if they do not yet exists

	tableCreateStatements := []struct {
		name string
		stmt string
	}{
		{name: "user", stmt: createUserTable},
		{name: "location", stmt: createLocationTable},
		{name: "dish", stmt: createDishTable},
		{name: "dish occurrences", stmt: createDishOccurrencesTable},
		{name: "user ratings", stmt: createDishRatingsTable},
	}

	for _, v := range tableCreateStatements {
		if _, err := db.Exec(v.stmt); err != nil {
			return nil, fmt.Errorf("failed to create table for %v : %v", v.name, err)
		}
	}

	return &MysqlRepo{db: db}, nil
}

func (m *MysqlRepo) finishTransaction(err error, tx *sql.Tx) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed with \"%v\" and rollback failed with \"%v\"", err, rollbackErr)
		}
		return err
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit : %w", err)
		}
		return nil
	}
}

// getOrCreateUser returns (id of user, if user was newly created, if error occurred)
func (m *MysqlRepo) getOrCreateUser(ctx context.Context, userEmail string) (int64, bool, error) {

	userID, err := m.getUser(ctx, userEmail)
	if err == nil {
		return userID, false, nil
	}
	//there was an error
	if !errors.Is(err, domain.ErrNotFound) {
		return 0, false, fmt.Errorf("getUser : %w", err)
	}
	//error was domain.ErrNotFound
	userID, err = m.createUser(ctx, userEmail, time.Now())
	if err != nil {
		return 0, false, fmt.Errorf("createUser : %w", err)
	}

	return userID, true, nil

}

func (m *MysqlRepo) getUser(ctx context.Context, userEmail string) (int64, error) {
	const rawStmt = `select id from users where email = ?`
	stmt, err := m.db.PrepareContext(ctx, rawStmt)
	if err != nil {
		return 0, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, userEmail)
	if err != nil {
		return 0, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	haveEntry := res.Next()
	if err := res.Err(); err != nil {
		return 0, fmt.Errorf("res.Err : %w", err)
	}
	if !haveEntry {
		return 0, domain.ErrNotFound
	}

	var id int64
	if err := res.Scan(&id); err != nil {
		return 0, fmt.Errorf("parsing user id : %w", err)
	}

	return id, nil
}

func (m *MysqlRepo) createUser(ctx context.Context, userEmail string, creationTime time.Time) (int64, error) {
	const rawStmt = `insert into users(email,created) VALUES (?,?)`
	stmt, err := m.db.PrepareContext(ctx, rawStmt)
	if err != nil {
		return 0, fmt.Errorf("PrepareContext for %v : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.ExecContext(ctx, userEmail, creationTime)
	if err != nil {
		return 0, fmt.Errorf("stmt %v : %w", rawStmt, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("LastInsertId : %w", err)
	}

	return id, nil
}

// getAllOccurrencesForDish fetches all occurences for the given dish using dbTX for its queries. Thus it can be embedded into transcations
func (m *MysqlRepo) getAllOccurrencesForDish(ctx context.Context, dbTX sqlContextPreparer, dishID int64) ([]time.Time, error) {
	const rawStmt = `select date from  dish_occurrences as do join dishes on do.dish_id = dishes.id where dishes.id = ?`
	stmt, err := dbTX.PrepareContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	occurrences := make([]time.Time, 0)
	for res.Next() {
		t := time.Time{}
		if err := res.Scan(&t); err != nil {
			return nil, fmt.Errorf("parsing result : %w", err)
		}

		//db contains utc time, convert back to local user time
		t = t.In(time.Local)

		occurrences = append(occurrences, t)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err : %w", err)
	}

	return occurrences, nil
}

func (m *MysqlRepo) getLocation(ctx context.Context, dbTX sqlContextPreparer, locationName string) (int64, error) {
	const rawStmt = `select id from locations where name = ?`
	stmt, err := dbTX.PrepareContext(ctx, rawStmt)
	if err != nil {
		return 0, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, locationName)
	if err != nil {
		return 0, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	haveEntry := res.Next()
	if err := res.Err(); err != nil {
		return 0, fmt.Errorf("res.Err : %w", err)
	}
	if !haveEntry {
		return 0, domain.ErrNotFound
	}

	var id int64
	if err := res.Scan(&id); err != nil {
		return 0, fmt.Errorf("parsing user id : %w", err)
	}

	return id, nil
}

func (m *MysqlRepo) createLocation(ctx context.Context, dbTX sqlContextPreparer, locationName string, creationTime time.Time) (int64, error) {
	const rawStmt = `insert into locations(name,created) VALUES (?,?)`
	stmt, err := dbTX.PrepareContext(ctx, rawStmt)
	if err != nil {
		return 0, fmt.Errorf("PrepareContext for %v : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.ExecContext(ctx, locationName, creationTime)
	if err != nil {
		return 0, fmt.Errorf("stmt %v : %w", rawStmt, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("LastInsertId : %w", err)
	}

	return id, nil
}

func (m *MysqlRepo) getOrCreateLocation(ctx context.Context, dbTX sqlContextPreparer, locationName string) (int64, bool, error) {
	locationID, err := m.getLocation(ctx, dbTX, locationName)
	if err == nil {
		log.Printf("Get location ")
		return locationID, false, nil
	}
	//there was an error
	if !errors.Is(err, domain.ErrNotFound) {
		return 0, false, fmt.Errorf("getLocation : %w", err)
	}
	//error was domain.ErrNotFound
	locationID, err = m.createLocation(ctx, dbTX, locationName, time.Now())
	if err != nil {
		return 0, false, fmt.Errorf("createLocation : %w", err)
	}

	return locationID, true, nil
}

func (m *MysqlRepo) createDishToday(ctx context.Context, dishName string, servedAt string) (dish *domain.Dish, dishID int64, createdLocation bool, err error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = m.finishTransaction(err, tx)
	}()

	dish = domain.NewDishToday(dishName, servedAt)

	//First create location if it does not yet exist
	var locationID int64
	locationID, createdLocation, err = m.getOrCreateLocation(ctx, tx, servedAt)
	if err != nil {
		err = fmt.Errorf("getOrCreateLocation : %w", err)
		return
	}
	//
	//Then create entry in dish table
	//

	const rawDishStmt = `insert into dishes(name,location_id) VALUES (?,?)`
	dishStmt, err := tx.PrepareContext(ctx, rawDishStmt)
	if err != nil {
		return nil, 0, false, fmt.Errorf("PrepareContext for %v : %w", rawDishStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase dishStmt : %v", err)
		}
	}(dishStmt)

	res, err := dishStmt.ExecContext(ctx, dish.Name, locationID)
	if err != nil {
		mysqlErr := &mysql.MySQLError{}
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == codeDuplicateEntry {
				log.Printf("failed to insert new dish : %v", err)
				err = alreadyExists
				return
			}
		}
		return nil, 0, false, fmt.Errorf("dishStmt %v : %w", rawDishStmt, err)

	}
	dishID, err = res.LastInsertId()
	if err != nil {
		return nil, 0, false, fmt.Errorf("LastInsertID : %w", err)
	}

	if err := dishStmt.Close(); err != nil {
		log.Printf("Failed to close dishStmt : %v", err)
	}
	//
	//Then create entry in dish_occurrences table. Cannot parallelize due to foreign key dependency
	//

	const rawOccurrenceStmt = `insert into dish_occurrences(dish_id,date) VALUES (?,?)`
	occurrenceStmt, err := tx.PrepareContext(ctx, rawOccurrenceStmt)
	if err != nil {
		return nil, 0, false, fmt.Errorf("PrepareContext for \"%v\" : %w", rawOccurrenceStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase occurrenceStmt : %v", err)
		}
	}(occurrenceStmt)

	occurrenceValues := dish.Occurrences()
	for i := range occurrenceValues {
		v := &occurrenceValues[i]
		if _, err := occurrenceStmt.ExecContext(ctx, dishID, v); err != nil {
			return nil, 0, false, fmt.Errorf("ExecContext on \"%v\" for dishID=%v,date=%v : %w", rawOccurrenceStmt, dishID, v, err)
		}
	}

	return dish, dishID, createdLocation, nil
}

func (m *MysqlRepo) GetOrCreateDish(ctx context.Context, dishName string, servedAt string) (*domain.Dish, bool, bool, int64, error) {
	dish, dishID, createdLocation, err := m.createDishToday(ctx, dishName, servedAt)
	if err == nil {
		return dish, true, createdLocation, dishID, nil
	}
	//got error
	if !errors.Is(err, alreadyExists) {
		return nil, false, false, 0, fmt.Errorf("failed to create dish : %w", err)
	}

	//if we are here, we got already exists error -> fetch dish
	dish, dishID, err = m.GetDishByName(ctx, dishName, servedAt)
	if err != nil {
		return nil, false, false, 0, fmt.Errorf("failed to get dish %w", err)
	}

	return dish, false, false, dishID, nil
}

func (m *MysqlRepo) getDishTableDTO(ctx context.Context, dbTX sqlContextPreparer, dishID int64) (*basicDishInfoDTO, error) {
	const rawStmt = `select dishes.name,locations.name from dishes join locations on dishes.location_id = locations.id 
                     where dishes.id = ?`
	stmt, err := dbTX.PrepareContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	haveEntry := res.Next()
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err : %w", err)
	}
	if !haveEntry {
		return nil, domain.ErrNotFound
	}

	var dto basicDishInfoDTO
	if err := res.Scan(&dto.Name, &dto.ServedAt); err != nil {
		return nil, fmt.Errorf("parsing db result : %v", err)
	}

	return &dto, nil
}

func (m *MysqlRepo) getDishID(ctx context.Context, dbTX sqlContextPreparer, dishName, servedAt string) (int64, error) {
	rawStmt := `select dishes.id from dishes join locations on dishes.location_id = locations.id where dishes.name = ?
 and locations.name = ?`
	stmt, err := dbTX.PrepareContext(ctx, rawStmt)
	if err != nil {
		return 0, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, dishName, servedAt)
	if err != nil {
		return 0, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	haveEntry := res.Next()
	if err := res.Err(); err != nil {
		return 0, fmt.Errorf("res.Err : %w", err)
	}
	if !haveEntry {
		return 0, domain.ErrNotFound
	}

	var id int64
	if err := res.Scan(&id); err != nil {
		return 0, fmt.Errorf("parsing user id : %w", err)
	}

	return id, nil
}

func (m *MysqlRepo) getDish(ctx context.Context, dbTX sqlContextPreparer, dishID int64) (*domain.Dish, error) {

	queryGroup, queryGroupCtx := errgroup.WithContext(ctx)

	//Get basic dish info
	var dishDTO *basicDishInfoDTO
	dishNotFound := false
	queryGroup.Go(func() error {
		var queryErr error
		dishDTO, queryErr = m.getDishTableDTO(queryGroupCtx, dbTX, dishID)
		if errors.Is(queryErr, domain.ErrNotFound) {
			dishNotFound = true
		}
		return queryErr
	})

	//Get occurrences
	var occurrences []time.Time
	queryGroup.Go(func() error {
		var queryErr error
		occurrences, queryErr = m.getAllOccurrencesForDish(queryGroupCtx, dbTX, dishID)
		return queryErr
	})

	//wait for sub queries
	err := queryGroup.Wait()
	if err != nil {
		if dishNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	dish := domain.NewDishFromDB(dishDTO.Name, dishDTO.ServedAt, occurrences)
	return dish, nil
}

func (m *MysqlRepo) getDishByName(ctx context.Context, dishName, servedAt string) (dish *domain.Dish, dishID int64, err error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = m.finishTransaction(err, tx)
	}()

	dishID, err = m.getDishID(ctx, tx, dishName, servedAt)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, 0, domain.ErrNotFound
		}
		dish = nil
		dishID = 0
		return
	}

	dish, err = m.getDish(ctx, tx, dishID)
	if err != nil {
		dish = nil
		dishID = 0
		return
	}

	return
}

func (m *MysqlRepo) GetDishByName(ctx context.Context, dishName, servedAt string) (dish *domain.Dish, dishID int64, err error) {

	//FIXME: not sure why m.getDishByName sometimes causes bad connection. Retrying seems to mitigate it but need to investigate
	retries := 3
	for retries > 0 {
		dish, dishID, err = m.getDishByName(ctx, dishName, servedAt)
		if err != nil {
			retries -= 1
			continue
		} else {
			break
		}
	}
	return
}

func (m *MysqlRepo) GetDishByID(ctx context.Context, dishID int64) (*domain.Dish, error) {
	dish, err := m.getDish(ctx, m.db, dishID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return dish, nil
}

func (m *MysqlRepo) GetDishByDate(ctx context.Context, when time.Time, optionalLocation *string) ([]int64, error) {

	//
	// Get dish matchingDishIDs matching query
	//

	rawStmt := `select dishes.id from dishes
    join dish_occurrences as do on dishes.id = do.dish_id
    join locations as loc on dishes.location_id = loc.id
    where date(do.date) = date(?)`

	if optionalLocation != nil {
		rawStmt += ` AND loc.name = ?`
	}

	stmt, err := m.db.PrepareContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	var res *sql.Rows
	if optionalLocation == nil {
		res, err = stmt.QueryContext(ctx, when)
	} else {
		res, err = stmt.QueryContext(ctx, when, *optionalLocation)
	}
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	matchingDishIDs := make([]int64, 0)

	for res.Next() {
		var id int64
		if err := res.Scan(&id); err != nil {
			return nil, fmt.Errorf("parsing result : %w", err)
		}
		matchingDishIDs = append(matchingDishIDs, id)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err : %w", err)
	}

	return matchingDishIDs, nil
}

func (m *MysqlRepo) UpdateMostRecentServing(ctx context.Context, dishID int64,
	updateFN func(currenMostRecent *time.Time) (*time.Time, error)) (err error) {

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = m.finishTransaction(err, tx)
	}()

	//get most recent occurrence date for dishID

	const rawGetStmt = `select date from dish_occurrences where dish_id = ? order by date desc limit 1`
	getStmt, err := tx.PrepareContext(ctx, rawGetStmt)
	if err != nil {
		err = fmt.Errorf("PrepareContext for \"%v\" failed with : %v", rawGetStmt, err)
		return
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to close getStmt : %v", err)
		}
	}(getStmt)

	res, err := getStmt.QueryContext(ctx, dishID)
	if err != nil {
		err = fmt.Errorf("QueryContext for \"%v\" : %w", rawGetStmt, err)
		return
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)
	haveEntry := res.Next()
	if resErr := res.Err(); resErr != nil {
		err = fmt.Errorf("res.Err : %w", resErr)
		return
	}
	var currentMostRecent *time.Time
	if !haveEntry {
		currentMostRecent = nil
	} else {
		t := time.Time{}
		if scanErr := res.Scan(&t); scanErr != nil {
			err = fmt.Errorf("scanning time : %w", scanErr)
			return
		}
		t = t.In(time.Local)
		currentMostRecent = &t
	}
	if err := res.Close(); err != nil {
		log.Printf("Failed to close rows : %v", err)
	}

	//execute update logic
	newMostRecent, err := updateFN(currentMostRecent)
	if err != nil {
		err = fmt.Errorf("updateFN failed : %v", err)
		return
	}
	//updateFN does not want to add new value
	if newMostRecent == nil {
		return
	}

	//if we come here, we have a new value to add
	const rawInsertStmt = `insert into dish_occurrences (dish_id,date) values (?,?)`
	insertStmt, err := tx.PrepareContext(ctx, rawInsertStmt)
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to close insertStmt : %v", err)
		}
	}(insertStmt)

	_, err = insertStmt.ExecContext(ctx, dishID, *newMostRecent)
	if err != nil {
		err = fmt.Errorf("ExecContext for \"%v\" : %w", rawGetStmt, err)
		return
	}
	return
}

func (m *MysqlRepo) GetAllDishIDs(ctx context.Context) ([]int64, error) {
	const rawStmt = `select id from dishes`
	res, err := m.db.QueryContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\": %v", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	ids := make([]int64, 0)

	for res.Next() {
		var id int64
		if err := res.Scan(&id); err != nil {
			return nil, fmt.Errorf("parsing result : %w", err)
		}
		ids = append(ids, id)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err : %w", err)
	}

	return ids, nil

}

func (m *MysqlRepo) GetRating(ctx context.Context, userEmail string, dishID int64) (*domain.DishRating, error) {
	const rawStmt = `select date,rating from dish_ratings as dr JOIN users on dr.user_id = users.id where
					email = ? and dish_id = ? `
	stmt, err := m.db.PrepareContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, userEmail, dishID)
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	haveEntry := res.Next()
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err : %w", err)
	}
	if !haveEntry {
		return nil, domain.ErrNotFound
	}

	t := time.Time{}
	rating := 0
	if err := res.Scan(&t, &rating); err != nil {
		return nil, fmt.Errorf("parsing GetRaing result : %w", err)
	}
	t = t.In(time.Local)

	dishRating, err := domain.NewDishRatingFromDB(userEmail, rating, t)
	if err != nil {
		return nil, fmt.Errorf("data validation by domain failed : %w", err)
	}

	return &dishRating, nil

}

func (m *MysqlRepo) SetOrCreateRating(ctx context.Context, userEmail string, dishID int64, rating domain.DishRating) (bool, error) {
	const rawStmt = `replace into dish_ratings(dish_id,user_id,date,rating) VALUES(?,?,?,?)`
	stmt, err := m.db.PrepareContext(ctx, rawStmt)

	if err != nil {
		return false, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	userID, _, err := m.getOrCreateUser(ctx, userEmail)
	if err != nil {
		return false, fmt.Errorf("getOrCreateUser for %v : %w", userEmail, err)
	}

	createdRating := false
	_, err = m.GetRating(ctx, userEmail, dishID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			createdRating = true
		} else {
			return false, fmt.Errorf("GetRating : %w", err)
		}
	}

	_, err = stmt.ExecContext(ctx, dishID, userID, rating.When, rating.Value)
	if err != nil {
		return false, fmt.Errorf("ExecContext : %w", err)
	}

	return createdRating, nil
}

func (m *MysqlRepo) GetAllRatingsForDish(ctx context.Context, dishID int64) (*domain.DishRatings, error) {

	//check that dish exist
	dish, err := m.GetDishByID(ctx, dishID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
	}

	//get all ratings for the dish
	const rawStmt = `select users.email, dr.date,dr.rating FROM dish_ratings as dr join users on dr.user_id = users.id
				  	 where dr.dish_id = ?`
	stmt, err := m.db.PrepareContext(ctx, rawStmt)
	if err != nil {
		return nil, fmt.Errorf("PrepareContext for \"%v\" : %v", rawStmt, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("Failed to clase stmt : %v", err)
		}
	}(stmt)

	res, err := stmt.QueryContext(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("QueryContext for \"%v\" : %w", rawStmt, err)
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			log.Printf("Failed to close rows : %v", err)
		}
	}(res)

	dishRatings := make([]domain.DishRating, 0)
	for res.Next() {
		userEmail := ""
		t := time.Time{}
		rating := 0
		if err := res.Scan(&userEmail, &t, &rating); err != nil {
			return nil, fmt.Errorf("parsing result : %w", err)
		}
		t = t.In(time.Local)

		dishRating, err := domain.NewDishRatingFromDB(userEmail, rating, t)
		if err != nil {
			return nil, fmt.Errorf("data validation by domain failed : %w", err)
		}
		dishRatings = append(dishRatings, dishRating)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("res.Err %w", err)
	}

	result := domain.NewDishRatings(*dish, dishRatings)
	return &result, nil

}

func (m *MysqlRepo) DropRepo(ctx context.Context) error {

	for _, table := range []string{"dish_ratings", "dish_occurrences", "users", "dishes", "locations"} {
		if _, err := m.db.ExecContext(ctx, fmt.Sprintf("drop table %v;", table)); err != nil {
			return fmt.Errorf("dropping %v : %v", table, err)
		}
	}
	return nil
}

func (m *MysqlRepo) Close() error {
	return m.db.Close()
}
