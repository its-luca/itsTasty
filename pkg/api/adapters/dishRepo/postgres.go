package dishRepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/volatiletech/null/v8"
	"itsTasty/pkg/api/adapters/dishRepo/sqlboilerPSQL"
	"itsTasty/pkg/api/domain"
	"log"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PostgresRepo struct {
	db              *sql.DB
	migrationSource migrate.MigrationSource
}

func NewPostgresRepo(db *sql.DB, migrationSource migrate.MigrationSource) (domain.DishRepo, error) {
	appliedMigrations, err := migrate.Exec(db, "postgres", migrationSource, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("failed to apply db migrations : %v", err)
	}
	if appliedMigrations != 0 {
		log.Printf("Applied %v migrations", appliedMigrations)
	}

	repo := &PostgresRepo{db: db, migrationSource: migrationSource}
	return repo, nil
}

// finishTransaction is a helper functions that performs a rollback if err != nil and commits the transaction otherwise
// the returned error includes potentials errors from a failed commit or rollback
func (p *PostgresRepo) finishTransaction(err error, tx *sql.Tx) error {
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

// getOrCreateUser is a helper functions that fetches the given user or creates it if it does not exist. Queries
// are executed on the given executor allowing to embed this into ongoing transactions
func (p *PostgresRepo) getOrCreateUser(ctx context.Context, userEmail string, executor boil.ContextExecutor) (*sqlboilerPSQL.User, error) {

	upsertUser := &sqlboilerPSQL.User{
		Email:   userEmail,
		Created: time.Now(),
	}
	err := upsertUser.Upsert(
		ctx,
		executor,
		false,
		nil,
		boil.Infer(), boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user : %v", err)
	}

	//N.B. Upsert does not seem to update the ID of the user in the no conflict case
	//Thus we need an additional query here to fetch the user

	dbUser, err := sqlboilerPSQL.Users(sqlboilerPSQL.UserWhere.Email.EQ(userEmail)).One(ctx, executor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user : %v", err)
	}

	return dbUser, nil

}

func (p *PostgresRepo) getOrCreateDish(ctx context.Context, dishName string, servedAt string) (resultDish *domain.Dish,
	isNewDish bool, isNewLocation bool, dishID int64, err error) {

	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//
	// Check if resultDish already exists
	//

	resultDish, dishID, err = p.getDishByName(tx, ctx, dishName, servedAt)
	//resultDish already exists, return it
	if err == nil {
		isNewLocation = false
		isNewDish = false
		return
	}

	if !errors.Is(err, domain.ErrNotFound) {
		err = fmt.Errorf("getOrCreateDish failed to check if resultDish exists: %v", err)
		return
	}

	//if we are here, the resultDish does not exist -> create it

	//
	//create location if it does not exist
	//

	var dbLocation *sqlboilerPSQL.Location
	//we update this in the if clause should we need to create a new location
	isNewLocation = false
	dbLocation, err = sqlboilerPSQL.Locations(sqlboilerPSQL.LocationWhere.Name.EQ(servedAt)).One(ctx, tx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("getOrCreateDish failed to check if location exists : %v", err)
			return
		}
		//if we are here, the location does not yet exist
		dbLocation = &sqlboilerPSQL.Location{
			Name:    servedAt,
			Created: time.Now(),
		}
		//N.B. that this updates dbLocation.ID to the newly generated ID
		err = dbLocation.Insert(ctx, tx, boil.Infer())
		if err != nil {
			err = fmt.Errorf("getOrCreateDish failed to insert new location : %v", err)
			return
		}
		isNewLocation = true
	}

	//
	//create resultDish
	//

	newDomainDish := domain.NewDishToday(dishName, servedAt)
	dbDish := sqlboilerPSQL.Dish{
		LocationID: dbLocation.ID,
		Name:       newDomainDish.Name,
	}
	err = dbDish.Insert(ctx, tx, boil.Infer())
	if err != nil {
		err = fmt.Errorf("getOrCreateDish failed to insert new resultDish : %v", err)
		return
	}
	isNewDish = true
	dishID = int64(dbDish.ID)

	//
	// add occurrence to resultDish
	//

	err = dbDish.AddDishOccurrences(ctx, tx, true, &sqlboilerPSQL.DishOccurrence{
		DishID: dbDish.ID,
		Date:   newDomainDish.Occurrences()[0],
	})
	if err != nil {
		err = fmt.Errorf("getOrCreateDish failed to insert resultDish occurence for newly created resultDish : %v", err)
		return
	}

	resultDish = domain.NewDishFromDB(dbDish.Name, servedAt, []time.Time{newDomainDish.Occurrences()[0]})

	return
}

func (p *PostgresRepo) GetOrCreateDish(ctx context.Context, dishName string, servedAt string) (*domain.Dish, bool, bool, int64, error) {
	return p.getOrCreateDish(ctx, dishName, servedAt)
}

func (p *PostgresRepo) getDishByName(exec boil.ContextExecutor, ctx context.Context, dishName, servedAt string) (dish *domain.Dish, dishID int64, err error) {

	//
	// Resolve servedAt to location id
	//

	dbLocation, err := sqlboilerPSQL.Locations(sqlboilerPSQL.LocationWhere.Name.EQ(servedAt)).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domain.ErrNotFound
		}
		return nil, 0, err
	}

	//
	// Query dish using location id
	//

	dbDish, err := sqlboilerPSQL.Dishes(
		sqlboilerPSQL.DishWhere.LocationID.EQ(dbLocation.ID),
		sqlboilerPSQL.DishWhere.Name.EQ(dishName),
		qm.Load(sqlboilerPSQL.DishRels.DishOccurrences), //eager load occurrences, as we want to iterate over all later on
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domain.ErrNotFound
		}
		return nil, 0, err
	}

	occurrences := make([]time.Time, 0, len(dbDish.R.DishOccurrences))
	for _, v := range dbDish.R.DishOccurrences {
		occurrences = append(occurrences, v.Date.In(time.Local))
	}

	dish = domain.NewDishFromDB(dbDish.Name, servedAt, occurrences)

	return dish, int64(dbDish.ID), nil

}

func (p *PostgresRepo) GetDishByName(ctx context.Context, dishName, servedAt string) (dish *domain.Dish, dishID int64, err error) {

	dbLocation, err := sqlboilerPSQL.Locations(sqlboilerPSQL.LocationWhere.Name.EQ(servedAt)).One(ctx, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domain.ErrNotFound
		}
		return nil, 0, err
	}

	dbDish, err := sqlboilerPSQL.Dishes(
		sqlboilerPSQL.DishWhere.LocationID.EQ(dbLocation.ID),
		sqlboilerPSQL.DishWhere.Name.EQ(dishName),
		qm.Load(sqlboilerPSQL.DishRels.DishOccurrences), //eager load occurrences, as we want to iterate over all later on
	).One(ctx, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domain.ErrNotFound
		}
		return nil, 0, err
	}

	occurrences := make([]time.Time, 0, len(dbDish.R.DishOccurrences))
	for _, v := range dbDish.R.DishOccurrences {
		occurrences = append(occurrences, v.Date)
	}

	dish = domain.NewDishFromDB(dbDish.Name, servedAt, occurrences)

	return dish, int64(dbDish.ID), nil

}

func (p *PostgresRepo) GetDishByID(ctx context.Context, dishID int64) (dish *domain.Dish, err error) {
	dbDish, err := sqlboilerPSQL.Dishes(
		sqlboilerPSQL.DishWhere.ID.EQ(int(dishID)),
		qm.Load(sqlboilerPSQL.DishRels.DishOccurrences), //eager load occurrences, as we want to iterate over all later on
		qm.Load(sqlboilerPSQL.DishRels.Location),
	).One(ctx, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	occurrences := make([]time.Time, 0, len(dbDish.R.DishOccurrences))
	for _, v := range dbDish.R.DishOccurrences {
		occurrences = append(occurrences, v.Date)
	}

	dish = domain.NewDishFromDB(dbDish.Name, dbDish.R.Location.Name, occurrences)

	return dish, nil
}

func (p *PostgresRepo) GetDishByDate(ctx context.Context, when time.Time, optionalLocation *string) ([]int64, error) {

	mods := []qm.QueryMod{
		//join dishes with dishOccurrences
		qm.InnerJoin(
			fmt.Sprintf("%s on %s = %s",
				sqlboilerPSQL.TableNames.DishOccurrences,
				sqlboilerPSQL.DishTableColumns.ID,
				sqlboilerPSQL.DishOccurrenceTableColumns.DishID)),
		//filter for "when"
		sqlboilerPSQL.DishOccurrenceWhere.Date.EQ(when),
	}

	if optionalLocation != nil {
		//join dishes with location
		mods = append(mods, qm.InnerJoin(
			fmt.Sprintf("%s on %s = %s",
				sqlboilerPSQL.TableNames.Locations,
				sqlboilerPSQL.DishTableColumns.LocationID,
				sqlboilerPSQL.LocationTableColumns.ID)))
		//filter for "optionalLocation"
		mods = append(mods, sqlboilerPSQL.LocationWhere.Name.EQ(*optionalLocation))
	}

	dbDishes, err := sqlboilerPSQL.Dishes(mods...).All(ctx, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]int64, 0), nil
		}
		return nil, fmt.Errorf("failed to fetch occurences for given timepoint : %v", err)
	}

	ids := make([]int64, 0, len(dbDishes))
	for _, v := range dbDishes {
		ids = append(ids, int64(v.ID))
	}

	return ids, nil

}

func (p *PostgresRepo) UpdateMostRecentServing(ctx context.Context, dishID int64, updateFN func(currenMostRecent *time.Time) (*time.Time, error)) (err error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//get most recent serving and lock it for update
	dbOccurence, err := sqlboilerPSQL.DishOccurrences(
		sqlboilerPSQL.DishOccurrenceWhere.DishID.EQ(int(dishID)),
		qm.OrderBy(sqlboilerPSQL.DishOccurrenceColumns.Date+" desc"),
		qm.For("update"),
	).One(ctx, tx)
	haveEntry := true
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("failed to fetch most recent occurence : %v", err)
			return

		}
		haveEntry = false
	}

	var oldMostRecentLocal *time.Time
	if haveEntry {
		t := dbOccurence.Date.Local()
		oldMostRecentLocal = &t
	}
	newMostRecent, err := updateFN(oldMostRecentLocal)
	if err != nil {
		err = fmt.Errorf("updateFN failed : %v", err)
		return
	}

	//updateFN does not want to add new value
	if newMostRecent == nil {
		return
	}

	//if we get here, we have a new value to add
	dbDishOccurrence := sqlboilerPSQL.DishOccurrence{
		DishID: int(dishID),
		Date:   *newMostRecent,
	}
	err = dbDishOccurrence.Insert(ctx, tx, boil.Infer())
	if err != nil {
		err = fmt.Errorf("failed to insert new occurence : %v", err)
		return
	}

	return
}

func (p *PostgresRepo) GetAllDishIDs(ctx context.Context) ([]int64, error) {
	dbDishes, err := sqlboilerPSQL.Dishes(qm.Select(sqlboilerPSQL.DishColumns.ID)).All(ctx, p.db)
	if err != nil {
		return nil, fmt.Errorf("failed to query dishes : %v", err)
	}

	ids := make([]int64, 0, len(dbDishes))
	for _, v := range dbDishes {
		ids = append(ids, int64(v.ID))
	}
	return ids, nil
}

func (p *PostgresRepo) GetRating(ctx context.Context, userEmail string, dishID int64) (*domain.DishRating, error) {

	dbRatings, err := sqlboilerPSQL.DishRatings(
		//join with users table
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			sqlboilerPSQL.TableNames.Users,
			sqlboilerPSQL.DishRatingTableColumns.UserID,
			sqlboilerPSQL.UserTableColumns.ID)),
		//filter for user email
		sqlboilerPSQL.UserWhere.Email.EQ(userEmail),
		//filter for dish id
		sqlboilerPSQL.DishRatingWhere.DishID.EQ(int(dishID)),
	).One(ctx, p.db)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query dish rating : %v", err)
	}

	domainDishRating, err := domain.NewDishRatingFromDB(userEmail, dbRatings.Rating, dbRatings.Date.Local())
	if err != nil {
		return nil, fmt.Errorf("failed to construct domain object from db data : %w", err)
	}

	return &domainDishRating, nil
}

func (p *PostgresRepo) setOrCreateRating(ctx context.Context, userEmail string, dishID int64, rating domain.DishRating) (isNew bool, err error) {
	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	dbUser, err := p.getOrCreateUser(ctx, userEmail, tx)
	if err != nil {
		return false, fmt.Errorf("failed to get or create user : %v", err)
	}

	_, err = sqlboilerPSQL.DishRatings(
		sqlboilerPSQL.DishRatingWhere.DishID.EQ(int(dishID)),
		sqlboilerPSQL.DishRatingWhere.UserID.EQ(dbUser.ID),
		qm.For("update"), //locks row so queries outside this transaction cannot modify it or delete it
	).One(ctx, tx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("failed to check if rating already exists : %w", err)
		}
		//have sql.ErrNoRows
		isNew = true
	}

	dbRating := sqlboilerPSQL.DishRating{
		DishID: int(dishID),
		UserID: dbUser.ID,
		Date:   rating.When,
		Rating: int(rating.Value),
	}
	if isNew {
		if err := dbRating.Insert(ctx, tx, boil.Infer()); err != nil {
			return isNew, fmt.Errorf("failed to insert new rating : %v", err)
		}
	} else {
		if _, err := dbRating.Update(ctx, tx, boil.Infer()); err != nil {
			return isNew, fmt.Errorf("failed to upate rating : %v", err)
		}
	}
	return isNew, nil

}

func (p *PostgresRepo) SetOrCreateRating(ctx context.Context, userEmail string, dishID int64, rating domain.DishRating) (bool, error) {

	return p.setOrCreateRating(ctx, userEmail, dishID, rating)
}

func (p *PostgresRepo) GetAllRatingsForDish(ctx context.Context, dishID int64) (*domain.DishRatings, error) {

	domDish, err := p.GetDishByID(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf(" failed to fetch dish : %w", err)
	}

	dbRatings, err := sqlboilerPSQL.DishRatings(
		sqlboilerPSQL.DishRatingWhere.DishID.EQ(int(dishID)),
		qm.Load(sqlboilerPSQL.DishRatingRels.User),
	).All(ctx, p.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dish ratings : %v", err)

	}

	domainDishRatings := make([]domain.DishRating, 0)
	for _, v := range dbRatings {
		domRating, err := domain.NewDishRatingFromDB(v.R.User.Email, v.Rating, v.Date.Local())
		if err != nil {
			return nil, fmt.Errorf("failed to construct domain object from db data : %w", err)
		}
		domainDishRatings = append(domainDishRatings, domRating)
	}

	res := domain.NewDishRatings(*domDish, domainDishRatings)
	return &res, nil
}

func (p *PostgresRepo) DropRepo(_ context.Context) error {
	_, err := migrate.Exec(p.db, "postgres", p.migrationSource, migrate.Down)
	if err != nil {
		return fmt.Errorf("failed to apply db migrations : %v", err)
	}
	return nil
}

func (p *PostgresRepo) Close() error {
	return p.db.Close()
}

func (p *PostgresRepo) CreateMergedDish(ctx context.Context, mergedDish *domain.MergedDish) (int64, error) {
	_, id, err := p.createMergedDish(ctx, mergedDish)
	if err != nil {
		return 0, fmt.Errorf("createMergedDish failed : %w", err)
	}

	return id, nil
}

func (p *PostgresRepo) createMergedDish(ctx context.Context, mergedDish *domain.MergedDish) (result *domain.MergedDish, id int64, err error) {

	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//get location id for location

	dbLocation, fetchErr := sqlboilerPSQL.Locations(
		sqlboilerPSQL.LocationWhere.Name.EQ(mergedDish.ServedAt),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("failed to fetch location with name \"%v\" : %w", mergedDish.ServedAt, err)
		return
	}

	//create merged dish
	dbMergedDish := &sqlboilerPSQL.MergedDish{
		Name:       mergedDish.Name,
		LocationID: dbLocation.ID,
	}
	err = dbMergedDish.Insert(ctx, tx, boil.Infer())
	if err != nil {
		err = fmt.Errorf("dbMergedDish.Insert failed : %v", err)
		return
	}

	//update db entries for dish1Name and dish2Name to add them to the merged dish

	//helper variable to set MergedAt to exact same time for dish1Name and dish2Name
	mergeTime := time.Now()

	for _, v := range mergedDish.GetCondensedDishNames() {
		dbDish, fetchErr := sqlboilerPSQL.Dishes(
			qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
				sqlboilerPSQL.TableNames.Locations,
				sqlboilerPSQL.DishTableColumns.LocationID,
				sqlboilerPSQL.LocationTableColumns.ID,
			)),
			sqlboilerPSQL.DishWhere.Name.EQ(v),
			sqlboilerPSQL.LocationWhere.Name.EQ(mergedDish.ServedAt),
			qm.For("update"),
		).One(ctx, tx)
		if fetchErr != nil {
			err = fmt.Errorf("sqlboilerPSQL.Dishes failed to query dish %v : %w", v, fetchErr)
			return
		}

		dbDish.MergedDishID.SetValid(dbMergedDish.ID)
		dbDish.MergedAt.SetValid(mergeTime)

		if _, err = dbDish.Update(ctx, tx, boil.Infer()); err != nil {
			err = fmt.Errorf("dbDish.Update failed to update dish (name:%v, id: %v) : %v", dbDish.Name, dbDish.ID, err)
			return
		}
	}

	//construct result
	result, id, err = p.getMergedDish(ctx, mergedDish.Name, mergedDish.ServedAt, tx)
	if err != nil {
		err = fmt.Errorf("p.getMergedDish failed for merged dish (id: %v, name: %v) : %w", dbMergedDish.ID, mergedDish.Name, err)
	}
	return

}

func (p *PostgresRepo) GetMergedDish(ctx context.Context, name, servedAt string) (*domain.MergedDish, int64, error) {
	mergedDish, id, err := p.getMergedDish(ctx, name, servedAt, p.db)
	if err != nil {
		return nil, 0, fmt.Errorf("getMergedDish failed (name: %v, servedAt: %v) : %w", name, servedAt, err)
	}
	return mergedDish, id, nil
}

func (p *PostgresRepo) getMergedDish(ctx context.Context, name, servedAt string, executor boil.ContextExecutor) (result *domain.MergedDish,
	id int64, err error) {
	//fetch merged dish from db
	dbMergedDish, err := sqlboilerPSQL.MergedDishes(
		sqlboilerPSQL.MergedDishWhere.Name.EQ(name),
		sqlboilerPSQL.LocationWhere.Name.EQ(servedAt),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			sqlboilerPSQL.TableNames.Locations,
			sqlboilerPSQL.LocationTableColumns.ID,
			sqlboilerPSQL.MergedDishTableColumns.LocationID)),
		qm.Load(sqlboilerPSQL.MergedDishRels.Dishes),
		qm.Load(sqlboilerPSQL.MergedDishRels.Location),
	).One(ctx, executor)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotFound
			return
		}
		err = fmt.Errorf("failed get query merged dish: %w", err)
		return
	}

	allCondensedDishNames := make([]string, 0)
	for _, v := range dbMergedDish.R.Dishes {
		allCondensedDishNames = append(allCondensedDishNames, v.Name)
	}

	dbMergedDish.R.GetLocation()
	if dbMergedDish.R.Location == nil {
		err = fmt.Errorf("dbMergedDish.R.Location was nil")
		return
	}
	result = domain.NewMergedDishFomDB(name, dbMergedDish.R.Location.Name, allCondensedDishNames)
	id = int64(dbMergedDish.ID)
	return
}

func (p *PostgresRepo) AddDishToMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) error {
	if err := p.addDishToMergedDish(ctx, mergedDishName, servedAt, dishName); err != nil {
		return fmt.Errorf("addDishToMergedDish failed : %w", err)
	}
	return nil
}

func (p *PostgresRepo) addDishToMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) (err error) {
	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//fetch merged dish to get id

	dbMergedDish, fetchErr := sqlboilerPSQL.MergedDishes(
		sqlboilerPSQL.MergedDishWhere.Name.EQ(mergedDishName),
		qm.InnerJoin(fmt.Sprintf("%s on %v = %v",
			sqlboilerPSQL.TableNames.Locations,
			sqlboilerPSQL.LocationTableColumns.ID,
			sqlboilerPSQL.MergedDishTableColumns.LocationID)),
		sqlboilerPSQL.LocationWhere.Name.EQ(servedAt),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("failed to fetch merged dish (name: %v, servedAt: %v) : %w",
			mergedDishName, servedAt, fetchErr)
		return
	}

	//use merged dish id to add dish
	dbDish, fetchErr := sqlboilerPSQL.Dishes(
		sqlboilerPSQL.DishWhere.Name.EQ(dishName),
		sqlboilerPSQL.DishWhere.LocationID.EQ(dbMergedDish.LocationID),
		qm.For("update"),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("sqlboilerPSQL.Dishes failed to query dish (name: %v, locationID %v ) : %w",
			dishName, dbMergedDish.LocationID, fetchErr)
		return
	}

	dbDish.MergedDishID.SetValid(dbMergedDish.ID)
	dbDish.MergedAt.SetValid(time.Now())
	if _, err = dbDish.Update(ctx, tx, boil.Infer()); err != nil {
		err = fmt.Errorf("dbDish.Update failed to update dish (name:%v, id: %v) : %v", dbDish.Name, dbDish.ID, err)
		return
	}

	return
}

func (p *PostgresRepo) RemoveDishFromMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) error {
	if err := p.removeDishFromMergedDish(ctx, mergedDishName, servedAt, dishName); err != nil {
		return fmt.Errorf("removeDishFromMergedDish failed : %w", err)
	}
	return nil
}

func (p *PostgresRepo) removeDishFromMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) (err error) {
	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//fetch merged
	dbMergedDish, fetchErr := sqlboilerPSQL.MergedDishes(
		sqlboilerPSQL.MergedDishWhere.Name.EQ(mergedDishName),
		qm.InnerJoin(fmt.Sprintf("%s on %v = %v",
			sqlboilerPSQL.TableNames.Locations,
			sqlboilerPSQL.LocationTableColumns.ID,
			sqlboilerPSQL.MergedDishTableColumns.LocationID)),
		sqlboilerPSQL.LocationWhere.Name.EQ(servedAt),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("failed to fetch merged dish (name: %v, servedAt: %v) : %w",
			mergedDishName, servedAt, fetchErr)
		return
	}

	dbDish, fetchErr := dbMergedDish.Dishes(
		sqlboilerPSQL.DishWhere.Name.EQ(dishName),
		sqlboilerPSQL.DishWhere.LocationID.EQ(dbMergedDish.LocationID),
		sqlboilerPSQL.DishWhere.MergedDishID.EQ(null.IntFrom(dbMergedDish.ID)),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("failed to fetch dish (name: %v, servedAt: %v) : %w",
			dishName, servedAt, fetchErr)
		return
	}

	if _, delErr := dbDish.Delete(ctx, tx); delErr != nil {
		err = fmt.Errorf("failed to delete dish (name: %v, servedAt :%v) : %w",
			dishName, servedAt, delErr)
		return
	}

	return
}

func (p *PostgresRepo) DeleteMergedDish(ctx context.Context, mergedDishName, servedAt string) error {
	if err := p.deleteMergedDish(ctx, mergedDishName, servedAt); err != nil {
		return fmt.Errorf("deleteMergedDish (name: %v, servedAt: %v) failed : %w", mergedDishName, servedAt, err)
	}
	return nil
}

func (p *PostgresRepo) deleteMergedDish(ctx context.Context, mergedDishName, servedAt string) (err error) {
	//
	//Create Transaction
	//

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("BeginTX : %w", err)
		return
	}
	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	//fetch merged
	dbMergedDish, fetchErr := sqlboilerPSQL.MergedDishes(
		sqlboilerPSQL.MergedDishWhere.Name.EQ(mergedDishName),
		qm.InnerJoin(fmt.Sprintf("%s on %v = %v",
			sqlboilerPSQL.TableNames.Locations,
			sqlboilerPSQL.LocationTableColumns.ID,
			sqlboilerPSQL.MergedDishTableColumns.LocationID)),
		sqlboilerPSQL.LocationWhere.Name.EQ(servedAt),
	).One(ctx, tx)
	if fetchErr != nil {
		err = fmt.Errorf("failed to fetch merged dish (name: %v, servedAt: %v) : %w",
			mergedDishName, servedAt, fetchErr)
		return
	}

	if _, delErr := dbMergedDish.Delete(ctx, tx); delErr != nil {
		err = fmt.Errorf("failed to delete merged dish : %w", err)
		return
	}

	return
}
