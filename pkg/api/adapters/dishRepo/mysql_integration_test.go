package dishRepo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"itsTasty/pkg/api/domain"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

const EnvMysqlTestDBListen = "TEST_MYSQL_DB_LISTEN"
const EnvMysqlTestDBUser = "TEST_MYSQL_DB_USER"
const EnvMysqlTestDBPW = "TEST_MYSQL_DB_PW"
const EnvMysqlTestDBName = "TEST_MYSQL_DB_NAME"

// mysqlIntegrationTestDB db connection for integration test. ALWAYS access via getMysqlIntegrationTestDB
var mysqlIntegrationTestDB *sql.DB = nil
var mysqlIntegrationTestDBLock sync.Mutex

// inMysqlIntegrationTestEnv returns true if the environment variables for the integration test environment are all set
func inMysqlIntegrationTestEnv() bool {
	return os.Getenv(EnvMysqlTestDBListen) != "" && os.Getenv(EnvMysqlTestDBUser) != "" && os.Getenv(EnvMysqlTestDBPW) != "" && os.Getenv(EnvMysqlTestDBName) != ""
}

// skipMysqlIntegrationTest calls t.Skipf with are standardized message
func skipMysqlIntegrationTest(t *testing.T) {
	t.Skipf("Skipping mysql integration test, as env vars %s, %s, %s and %s are not set", EnvMysqlTestDBListen, EnvMysqlTestDBUser, EnvMysqlTestDBPW, EnvMysqlTestDBName)
}

// getMysqlIntegrationTestDB returns a db handle for use in integration test. Multiple test may receive the same handle.
// safe to call concurrently
func getMysqlIntegrationTestDB() (*sql.DB, error) {
	mysqlIntegrationTestDBLock.Lock()
	defer mysqlIntegrationTestDBLock.Unlock()

	if mysqlIntegrationTestDB != nil {
		return mysqlIntegrationTestDB, nil
	}

	buildDSN := func(user, pw, url, dbName string) string {
		return fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", user, pw, url, dbName)
	}
	err := mysql.SetLogger(log.New(ioutil.Discard, "", 0))
	if err != nil {
		return nil, fmt.Errorf("failed to set discard logger mysql")
	}

	db, err := sql.Open("mysql", buildDSN(os.Getenv(EnvMysqlTestDBUser), os.Getenv(EnvMysqlTestDBPW), os.Getenv(EnvMysqlTestDBListen), os.Getenv(EnvMysqlTestDBName)))
	if err != nil {
		return nil, fmt.Errorf("sql.Open dsn %v : %v", buildDSN(os.Getenv(EnvMysqlTestDBUser), os.Getenv(EnvMysqlTestDBPW), os.Getenv(EnvMysqlTestDBListen), os.Getenv(EnvMysqlTestDBName)), err)
	}

	retries := 10
	connected := false
	var lastErr error = nil
	for retries > 0 && !connected {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := db.PingContext(pingCtx)
		if err != nil {
			lastErr = err
			retries -= 1
			time.Sleep(3 * time.Second)
		} else {
			connected = true
			lastErr = nil
		}
		pingCancel()
	}

	if !connected {
		return nil, fmt.Errorf("error connecting to db : %v", lastErr)
	}

	mysqlIntegrationTestDB = db
	return db, nil
}

// TestNewDblpRepo simply tests if NewDblpRepo runs without errors, indicating that all the table creation sql code
// runs fine
func TestNewDblpRepo(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	//
	// Run Test
	//

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}

	if err := repo.DropRepo(context.Background()); err != nil {
		t.Errorf("Error cleanup up repo : %v", err)
	}
}

func TestMysqlRepo_getOrCreateUser(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//
	// Run Test
	//

	const userEmail = "test@user.ser"
	const userEmail2 = "test2@user.ser"

	// Running for the first time, we expect a new user to be created
	wantUserID, isNew, err := repo.getOrCreateUser(context.Background(), userEmail)
	require.NoError(t, err)
	require.True(t, isNew)

	// Running for the second time, we expect to get the same id and "isNew" to be false

	gotUserID, isNew, err := repo.getOrCreateUser(context.Background(), userEmail)
	require.NoError(t, err)
	require.False(t, isNew)
	require.Equal(t, wantUserID, gotUserID)

	// Using a different email, we again expect a new user

	gotUserID2, isNew, err := repo.getOrCreateUser(context.Background(), userEmail2)
	require.NoError(t, err)
	require.True(t, isNew)
	require.NotEqual(t, gotUserID, gotUserID2)
}

func TestMysqlRepo_GetOrCreateDish_CreateAndQuery(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//
	// Run Test
	//

	wantDish := domain.NewDishToday("Test Dish A", "testLocation")
	//Create new dish
	gotNewDish, createdNewDish, createdNewLocation, gotNewDishID, err := repo.GetOrCreateDish(context.Background(), wantDish.Name, wantDish.ServedAt)
	require.NoError(t, err)
	require.True(t, createdNewDish)
	require.True(t, createdNewLocation)

	require.Equal(t, wantDish.Name, gotNewDish.Name)
	require.Equal(t, wantDish.ServedAt, gotNewDish.ServedAt)
	//We cannot directly check the time as it is generated based on time.Now
	require.Equal(t, 1, len(gotNewDish.Occurrences()))
	require.True(t, gotNewDish.Occurrences()[0].Sub(time.Now()) < 10*time.Second)

	//Call again with same dish name -> expect to get dish instead of creating one
	gotExistingDish, createdNewDish, createdNewLocation, gotExistingDishID, err := repo.GetOrCreateDish(context.Background(), wantDish.Name, wantDish.ServedAt)
	require.NoError(t, err)
	require.False(t, createdNewDish)
	require.False(t, createdNewLocation)
	require.Equal(t, gotNewDishID, gotExistingDishID)
	require.Equal(t, gotNewDish, gotExistingDish)

	//Call GetDishByName  on existing dish
	gotExistingDish, gotExistingDishID, err = repo.GetDishByName(context.Background(), wantDish.Name, wantDish.ServedAt)
	require.NoError(t, err)
	require.Equal(t, gotNewDishID, gotExistingDishID)
	require.Equal(t, gotNewDish, gotExistingDish)

	//Call GetDishByID  on existing dish
	gotExistingDish, err = repo.GetDishByID(context.Background(), gotNewDishID)
	require.NoError(t, err)
	require.Equal(t, gotNewDish, gotExistingDish)

	//Call GetOrCreateDish on dish with same name but different location
	_, createdNewDish, createdNewLocation, gotDishID, err := repo.GetOrCreateDish(context.Background(), wantDish.Name, "newLocation")
	require.NoError(t, err)
	require.True(t, createdNewDish)
	require.True(t, createdNewLocation)
	require.NotEqual(t, gotDishID, gotNewDish)
}

func TestMysqlRepo_GetOrCreateDish_CheckNotFoundError(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//
	// Run Test
	//

	_, _, err = repo.GetDishByName(context.Background(), "does not exist", "someLocation")
	require.Equal(t, err, domain.ErrNotFound)

	_, err = repo.GetDishByID(context.Background(), 42)
	require.Equal(t, err, domain.ErrNotFound)
}

func TestMysqlRepo_GetAllDishIDs(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//
	// Run Test
	//

	//Initially there should be no dish
	ids, err := repo.GetAllDishIDs(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, len(ids))

	//Create dish and query again
	_, _, _, dishID, err := repo.GetOrCreateDish(context.Background(), "testDish", "testLocation")
	require.NoError(t, err)
	ids, err = repo.GetAllDishIDs(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{dishID}, ids)

}

func Test_SetOrCreateRating(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//add dish
	const sampleDishName = "sampleDish"
	const sampleUserEmail = "test@use.er"
	_, isNewDish, isNewLocation, sampleDishID, err := repo.GetOrCreateDish(context.Background(), sampleDishName, "someLocation")
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	//
	// Run Test
	//

	//Note: Mysql does not seem to retain nano seconds. Thus, we have to round or time values to second
	//to make "got want" comparisons work

	//Initial call should create new rating
	initialDishRating := domain.NewDishRating(sampleUserEmail, domain.FiveStars, time.Now().Round(time.Second))
	isNew, err := repo.SetOrCreateRating(context.Background(), sampleUserEmail, sampleDishID, initialDishRating)
	require.NoError(t, err)
	require.True(t, isNew)

	//Get the newly created rating
	dishRating, err := repo.GetRating(context.Background(), sampleUserEmail, sampleDishID)
	require.NoError(t, err)
	require.Equal(t, initialDishRating, *dishRating)

	//Second call should only change the rating
	updatedDishRating := domain.NewDishRating(sampleUserEmail, domain.OneStar, time.Now().Add(2*time.Second).Round(time.Second))
	isNew, err = repo.SetOrCreateRating(context.Background(), sampleUserEmail, sampleDishID, updatedDishRating)
	require.NoError(t, err)
	require.False(t, isNew)

	//Get the updated rating
	dishRating, err = repo.GetRating(context.Background(), sampleUserEmail, sampleDishID)
	require.NoError(t, err)
	require.Equal(t, updatedDishRating, *dishRating)

	//Use GetAll to also check that there is only one rating
	ratings, err := repo.GetAllRatingsForDish(context.Background(), sampleDishID)
	require.NoError(t, err)
	count := 0
	for _, v := range ratings.Ratings() {
		count += v
	}
	require.Equal(t, 1, count)
}

func TestMysqlRepo_UpdateMostRecentServing(t *testing.T) {
	//
	//Build test env
	//
	if !inMysqlIntegrationTestEnv() {
		skipMysqlIntegrationTest(t)
	}
	db, err := getMysqlIntegrationTestDB()
	if err != nil {
		t.Fatalf("getMysqlIntegrationTestDB : %v", err)
	}

	repo, err := NewMysqlRepo(db)
	if err != nil {
		t.Errorf("NewDblpRepo failed with %v", err)
	}
	defer func() {
		if err := repo.DropRepo(context.Background()); err != nil {
			t.Errorf("Error cleanup up repo : %v", err)
		}
	}()

	//add dish
	const sampleDishName = "sampleDish"
	const sampleUserEmail = "test@use.er"
	_, isNewDish, isNewLocation, sampleDishID, err := repo.GetOrCreateDish(context.Background(), sampleDishName, "someLocation")
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	//
	// Run Test
	//

	//Initially there should be one serving for today, created by GetOrCreateDish
	err = repo.UpdateMostRecentServing(context.Background(), sampleDishID, func(currenMostRecent *time.Time) (*time.Time, error) {
		require.True(t, domain.OnSameDay(time.Now(), *currenMostRecent))
		return nil, nil
	})
	require.NoError(t, err)

	//There still should be no serving, but now we crate one
	wantServingDate := domain.NowWithDayPrecision().Add(24 * time.Hour)
	err = repo.UpdateMostRecentServing(context.Background(), sampleDishID, func(currenMostRecent *time.Time) (*time.Time, error) {
		require.True(t, domain.OnSameDay(time.Now(), *currenMostRecent))
		return &wantServingDate, nil
	})
	require.NoError(t, err)

	//Check that the serving has the expected value
	err = repo.UpdateMostRecentServing(context.Background(), sampleDishID, func(currenMostRecent *time.Time) (*time.Time, error) {
		require.Equal(t, wantServingDate, *currenMostRecent)
		return nil, nil
	})
	require.NoError(t, err)
}
