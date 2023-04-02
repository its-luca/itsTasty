package dishRepo

import (
	"context"
	"itsTasty/pkg/api/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type factoryCleanupFunc func() error
type repoFactory func() (domain.DishRepo, factoryCleanupFunc, error)
type dbTestFunc func(t *testing.T, repo domain.DishRepo)

// roundTimeToDBResolution is a helper that rounds down the time precision, as the database
// does not seem to retain nanoseconds. This helps to compare time values in tests
func roundTimeToDBResolution(t time.Time) time.Time {
	return t.Round(time.Second)
}

type commonDbTest struct {
	Name     string
	TestFunc dbTestFunc
}

func runCommonDbTests(t *testing.T, factory repoFactory) {

	tests := []commonDbTest{
		{
			Name:     "GetOrCreateDish_CreateAndQuery",
			TestFunc: testRepo_GetOrCreateDish_CreateAndQuery,
		},
		{
			Name:     "GetOrCreateDish_CheckNotFoundError",
			TestFunc: testRepo_GetOrCreateDish_CheckNotFoundError,
		},
		{
			Name:     "GetAllDishIDs",
			TestFunc: testRepo_GetAllDishIDs,
		},
		{
			Name:     "UpdateMostRecentServing",
			TestFunc: testRepo_UpdateMostRecentServing,
		},
		{
			Name:     "GetDishByDate",
			TestFunc: testRepo_GetDishByDate,
		},
		{
			Name:     "UpdateMostRecentRating_and_GetRatings",
			TestFunc: test_UpdateMostRecentRating_GetRatings,
		},
		{
			Name:     "CreateMergedDish",
			TestFunc: testPostgresRepo_CreateMergedDish,
		},
		{
			Name:     "AddDishToMergedDish",
			TestFunc: testPostgresRepo_AddDishToMergedDish,
		},
		{
			Name:     "RemoveDishFromMergedDish",
			TestFunc: testPostgresRepo_RemoveDishFromMergedDish,
		},
		{
			Name:     "DeleteMergedDish",
			TestFunc: testPostgresRepo_DeleteMergedDish,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			repo, cleanup, err := factory()
			require.NoError(t, err)
			defer func() {
				if err := cleanup(); err != nil {
					t.Fatalf("Cleanup failed : %v", err)
				}
			}()
			defer func() {
				err = repo.DropRepo(context.Background())
				require.NoError(t, err)
			}()

			test.TestFunc(t, repo)
		})
	}
}

func testRepo_GetOrCreateDish_CreateAndQuery(t *testing.T, repo domain.DishRepo) {

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

func testRepo_GetOrCreateDish_CheckNotFoundError(t *testing.T, repo domain.DishRepo) {

	_, _, err := repo.GetDishByName(context.Background(), "does not exist", "someLocation")
	require.Equal(t, err, domain.ErrNotFound)

	_, err = repo.GetDishByID(context.Background(), 42)
	require.Equal(t, err, domain.ErrNotFound)
}

func testRepo_GetAllDishIDs(t *testing.T, repo domain.DishRepo) {
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

func test_UpdateMostRecentRating_GetRatings(t *testing.T, repo domain.DishRepo) {
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

	//Note: Mysql does not seem to retain nanoseconds. Thus, we have to round our time values to second
	//to make "got want" comparisons work

	timeFirstRating := roundTimeToDBResolution(time.Now())
	timeSecondRating := timeFirstRating.Add(24 * time.Hour)
	firstRating := domain.NewDishRating(sampleUserEmail, domain.FiveStars, timeFirstRating)
	err = repo.CreateOrUpdateRating(context.Background(), sampleUserEmail, sampleDishID,
		func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {

			require.Nilf(t, currentRating, "on first call to CreateOrUpdateRating \"currentRating\" should be nil")

			updatedRating = &firstRating
			createNew = true
			err = nil
			return
		})
	require.NoError(t, err)

	//Get the newly created rating and check values
	dishRatings, err := repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, false)
	require.NoError(t, err)
	require.Equal(t, []domain.DishRating{firstRating}, dishRatings)

	//Create a second rating
	secondRating := domain.NewDishRating(sampleUserEmail, domain.OneStar, timeSecondRating)
	err = repo.CreateOrUpdateRating(context.Background(), sampleUserEmail, sampleDishID,
		func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {

			require.Equalf(t, firstRating, *currentRating, "on second call to CreateOrUpdateRating"+
				"\"currentRating\" have the value of \"firstRating\"")

			updatedRating = &secondRating
			createNew = true
			err = nil
			return
		})
	require.NoError(t, err)

	//GetRatings should return both ratings if "onlyMostRecent" is set to false
	gotRatings, err := repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, false)
	require.NoError(t, err)
	require.Equal(t, []domain.DishRating{firstRating, secondRating}, gotRatings)

	//GetRatings should return ONLY MOST RECENT (although there are now two ratings) since onyMostRecent is set to true
	gotRatings, err = repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, true)
	require.NoError(t, err)
	require.Equal(t, []domain.DishRating{secondRating}, gotRatings)

	//Update the most recent rating
	updatedSecondRating := domain.NewDishRating(sampleUserEmail, domain.ThreeStars, timeSecondRating.Add(10*time.Minute))
	err = repo.CreateOrUpdateRating(context.Background(), sampleUserEmail, sampleDishID,
		func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {

			updatedRating = &updatedSecondRating
			createNew = false
			err = nil
			return
		})
	require.NoError(t, err)

	//Fetch most recent rating and check for updated values
	gotRatings, err = repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, true)
	require.NoError(t, err)
	require.Equal(t, []domain.DishRating{updatedSecondRating}, gotRatings)

	//Check that no date is updated if we return nil in updateFN
	err = repo.CreateOrUpdateRating(context.Background(), sampleUserEmail, sampleDishID,
		func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {

			updatedRating = nil
			createNew = false
			err = nil
			return
		})
	require.NoError(t, err)

	//Fetch most recent rating and check that values did not change
	gotRatings, err = repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, true)
	require.NoError(t, err)
	require.Equal(t, []domain.DishRating{updatedSecondRating}, gotRatings)

	//Check that there are only two ratings in total
	gotRatings, err = repo.GetRatings(context.Background(), sampleUserEmail, sampleDishID, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(gotRatings))

}

func testRepo_UpdateMostRecentServing(t *testing.T, repo domain.DishRepo) {

	//add dish
	const sampleDishName = "sampleDish"
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
	wantServingDate := domain.TruncateToDayPrecision(time.Now()).Add(24 * time.Hour)
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

func testRepo_GetDishByDate(t *testing.T, repo domain.DishRepo) {
	locationA := "locationA"
	locationB := "locationB"

	//add test dishes
	_, isNewDish, isNewLocation, wantDish1LocationA, err := repo.GetOrCreateDish(context.Background(), "dish1LocationA", locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	_, isNewDish, isNewLocation, wantDish2LocationA, err := repo.GetOrCreateDish(context.Background(), "dish2LocationA", locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	_, isNewDish, isNewLocation, wantDish1LocationB, err := repo.GetOrCreateDish(context.Background(), "dish1LocationB", locationB)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	//
	// Run Test
	//

	//Get all dishes without filtering for location

	gotDishIDs, err := repo.GetDishByDate(context.Background(), domain.TruncateToDayPrecision(time.Now()), nil)
	require.NoError(t, err)
	wantDishIDs := []int64{wantDish1LocationA, wantDish2LocationA, wantDish1LocationB}
	require.ElementsMatch(t, wantDishIDs, gotDishIDs)

	//Get all dishes from locationA
	gotDishIDs, err = repo.GetDishByDate(context.Background(), domain.TruncateToDayPrecision(time.Now()), &locationA)
	require.NoError(t, err)
	wantDishIDs = []int64{wantDish1LocationA, wantDish2LocationA}
	require.ElementsMatch(t, wantDishIDs, gotDishIDs)

	//search for non-existing location
	nonExistingLocation := "nonExistingLocation"
	gotDishIDs, err = repo.GetDishByDate(context.Background(), domain.TruncateToDayPrecision(time.Now()), &nonExistingLocation)
	require.NoError(t, err)
	require.Equal(t, 0, len(gotDishIDs))

	//search for non-existing time
	gotDishIDs, err = repo.GetDishByDate(context.Background(), domain.TruncateToDayPrecision(time.Now()).Add(24*time.Hour), nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(gotDishIDs))
}

func testPostgresRepo_CreateMergedDish(t *testing.T, repo domain.DishRepo) {

	//Setup : Create 2 dishes on same location
	const locationA = "locationA"
	const dishAName = "dishA"
	const dishBName = "dishB"

	_, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishAName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	_, isNewDish, isNewLocation, _, err = repo.GetOrCreateDish(context.Background(), dishBName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	//
	// Run Test : Create merged dish, fetch dish, compare
	//
	const mergedDishName = "dishABMerged"
	wantDishA := domain.NewDishToday(dishAName, locationA)
	wantDishB := domain.NewDishToday(dishBName, locationA)

	//Create new merged dish and insert it into db

	mergedDish, err := domain.NewMergedDish(mergedDishName, wantDishA, wantDishB, []*domain.Dish{})
	require.NoError(t, err)

	wantMergedDishID, err := repo.CreateMergedDish(context.Background(), mergedDish)
	require.NoError(t, err)

	//Fetch merged dish from db and compare
	gotMergedDish, gotMergedDishID, err := repo.GetMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.NoError(t, err)
	require.Equalf(t, wantMergedDishID, gotMergedDishID, "fetched mergedDishID does not match")
	require.Equalf(t, mergedDish, gotMergedDish, "fetched mergedDish does not match")

}

func testPostgresRepo_AddDishToMergedDish(t *testing.T, repo domain.DishRepo) {

	//Setup : Create 3 dishes on same location
	const locationA = "locationA"
	const dishAName = "dishA"
	const dishBName = "dishB"
	const dishCName = "dishC"

	_, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishAName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	_, isNewDish, isNewLocation, _, err = repo.GetOrCreateDish(context.Background(), dishBName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	dishC, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishCName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	//
	// Run Test : Create merged dish, add additional dish to it, fetch and check
	//
	const mergedDishName = "dishABMerged"
	wantDishA := domain.NewDishToday(dishAName, locationA)
	wantDishB := domain.NewDishToday(dishBName, locationA)

	//Create new merged dish and insert it into db

	mergedDish, err := domain.NewMergedDish(mergedDishName, wantDishA, wantDishB, []*domain.Dish{})
	require.NoError(t, err)

	wantMergedDishID, err := repo.CreateMergedDish(context.Background(), mergedDish)
	require.NoError(t, err)

	//add additional dish

	err = repo.UpdateMergedDishByID(context.Background(), wantMergedDishID, func(current *domain.MergedDish) (*domain.MergedDish, error) {
		if err := current.AddDish(dishC); err != nil {
			return nil, err
		}

		return current, nil
	})
	require.NoError(t, err)

	//Fetch merged dish from db and compare
	gotMergedDish, gotMergedDishID, err := repo.GetMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.NoError(t, err)
	require.Equalf(t, wantMergedDishID, gotMergedDishID, "fetched mergedDishID does not match")
	require.ElementsMatch(t, []string{dishAName, dishBName, dishCName}, gotMergedDish.GetCondensedDishNames(),
		"fetched mergedDish contain expected dishes")

}

func testPostgresRepo_RemoveDishFromMergedDish(t *testing.T, repo domain.DishRepo) {

	//Setup : Create 3 dishes on same location
	const locationA = "locationA"
	const dishAName = "dishA"
	const dishBName = "dishB"
	const dishCName = "dishC"

	_, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishAName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	_, isNewDish, isNewLocation, _, err = repo.GetOrCreateDish(context.Background(), dishBName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	dishC, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishCName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	//
	// Run Test : Create merged dish with 3 dishes, remove one dish, fetch and check
	//
	const mergedDishName = "dishABMerged"
	wantDishA := domain.NewDishToday(dishAName, locationA)
	wantDishB := domain.NewDishToday(dishBName, locationA)
	wantDishC := domain.NewDishToday(dishCName, locationA)

	//Create new merged dish and insert it into db

	mergedDish, err := domain.NewMergedDish(mergedDishName, wantDishA, wantDishB, []*domain.Dish{wantDishC})
	require.NoError(t, err)

	wantMergedDishID, err := repo.CreateMergedDish(context.Background(), mergedDish)
	require.NoError(t, err)

	//remove dish C

	err = repo.UpdateMergedDishByID(context.Background(), wantMergedDishID, func(current *domain.MergedDish) (*domain.MergedDish, error) {
		if err := current.RemoveDish(dishC); err != nil {
			return nil, err
		}
		return current, nil
	})
	require.NoError(t, err)

	//Fetch merged dish from db and compare
	gotMergedDish, gotMergedDishID, err := repo.GetMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.NoError(t, err)
	require.Equalf(t, wantMergedDishID, gotMergedDishID, "fetched mergedDishID does not match")
	require.ElementsMatch(t, []string{dishAName, dishBName}, gotMergedDish.GetCondensedDishNames(),
		"fetched mergedDish contain expected dishes")

}

func testPostgresRepo_DeleteMergedDish(t *testing.T, repo domain.DishRepo) {

	//Setup : Create 2 dishes on same location
	const locationA = "locationA"
	const dishAName = "dishA"
	const dishBName = "dishB"

	_, isNewDish, isNewLocation, _, err := repo.GetOrCreateDish(context.Background(), dishAName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	_, isNewDish, isNewLocation, _, err = repo.GetOrCreateDish(context.Background(), dishBName, locationA)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.False(t, isNewLocation)

	//
	// Run Test : Create merged dish, fetch dish, delete dish, fetch again
	//
	const mergedDishName = "dishABMerged"
	wantDishA := domain.NewDishToday(dishAName, locationA)
	wantDishB := domain.NewDishToday(dishBName, locationA)

	//Create new merged dish and insert it into db

	mergedDish, err := domain.NewMergedDish(mergedDishName, wantDishA, wantDishB, []*domain.Dish{})
	require.NoError(t, err)

	wantMergedDishID, err := repo.CreateMergedDish(context.Background(), mergedDish)
	require.NoError(t, err)

	//Fetch merged dish from db and compare
	gotMergedDish, gotMergedDishID, err := repo.GetMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.NoError(t, err)
	require.Equalf(t, wantMergedDishID, gotMergedDishID, "fetched mergedDishID does not match")
	require.Equalf(t, mergedDish, gotMergedDish, "fetched mergedDish does not match")

	//Delete merged dish
	err = repo.DeleteMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.NoError(t, err)

	//Try to fetch merged dish again. Expecting not found error
	_, _, err = repo.GetMergedDish(context.Background(), mergedDish.Name, mergedDish.ServedAt)
	require.ErrorIsf(t, err, domain.ErrNotFound, "expected not found error since we deleted the dish")
}
