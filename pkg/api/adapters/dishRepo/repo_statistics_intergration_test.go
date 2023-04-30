package dishRepo

import (
	"context"
	"github.com/stretchr/testify/require"
	"itsTasty/pkg/api/domain"
	"testing"
	"time"
)

func testStatistics_GetAllUsers(t *testing.T, repo *PostgresRepo) {

	ctx := context.Background()
	wantDishName := "testDish"
	wantDishLocation := "testLocation"
	wantUser1 := "user1@testuser"
	wantUser2 := "user2@testuser"

	//create dish and one rating from each users. We need the ratings as they create the users in the db
	_, isNewDish, isNewLocation, dishID, err := repo.GetOrCreateDish(ctx, wantDishName, wantDishLocation)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	err = repo.CreateOrUpdateRating(ctx, wantUser1, dishID, func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {
		rating := domain.NewDishRating(wantUser1, domain.ThreeStars, roundTimeToDBResolution(time.Now()))
		return &rating, true, nil
	})
	require.NoError(t, err)
	err = repo.CreateOrUpdateRating(ctx, wantUser2, dishID, func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {
		rating := domain.NewDishRating(wantUser2, domain.ThreeStars, roundTimeToDBResolution(time.Now()))
		return &rating, true, nil
	})
	require.NoError(t, err)

	//check that two users are returned
	gotUsers, err := repo.GetAllUsers(ctx)
	wantUsers := []domain.User{
		{Email: wantUser1},
		{Email: wantUser2},
	}
	require.NoError(t, err)
	require.ElementsMatch(t, wantUsers, gotUsers)
}

func testStatistics_GetAllRatingsForDate(t *testing.T, repo *PostgresRepo) {

	ctx := context.Background()
	wantDishName := "testDish"
	wantDishLocation := "testLocation"

	date1 := domain.NewDayPrecisionTime(roundTimeToDBResolution(time.Now()))
	date2 := domain.NewDayPrecisionTime(roundTimeToDBResolution(time.Now().Add(24 * time.Hour)))

	wantUser1 := "user1@testuser"
	wantUser2 := "user2@testuser"
	wantUser3 := "user3@testuser"

	wantRating1Date1 := domain.NewDishRating(wantUser1, domain.ThreeStars, date1.Time)
	wantRating2Date1 := domain.NewDishRating(wantUser2, domain.ThreeStars, date1.Time)
	wantRating1Date2 := domain.NewDishRating(wantUser3, domain.ThreeStars, date2.Time)

	//create dish and three ratings, two on date1 one on date2
	_, isNewDish, isNewLocation, dishID, err := repo.GetOrCreateDish(ctx, wantDishName, wantDishLocation)
	require.NoError(t, err)
	require.True(t, isNewDish)
	require.True(t, isNewLocation)

	err = repo.CreateOrUpdateRating(ctx, wantUser1, dishID, func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {
		return &wantRating1Date1, true, nil
	})
	require.NoError(t, err)

	err = repo.CreateOrUpdateRating(ctx, wantUser2, dishID, func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {
		return &wantRating2Date1, true, nil
	})
	require.NoError(t, err)

	err = repo.CreateOrUpdateRating(ctx, wantUser3, dishID, func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {
		return &wantRating1Date2, true, nil
	})
	require.NoError(t, err)

	//check that only the rating from the selected date is returned
	gotRatings, err := repo.GetAllRatingsForDate(ctx, date1)
	wantRatings := []domain.DishRating{wantRating1Date1, wantRating2Date1}
	require.NoError(t, err)
	require.ElementsMatch(t, wantRatings, gotRatings)
}
