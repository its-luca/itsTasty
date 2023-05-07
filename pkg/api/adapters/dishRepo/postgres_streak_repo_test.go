package dishRepo

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"itsTasty/pkg/api/domain"
	"testing"
	"time"
)

func testStreak_Create_Get_Update(t *testing.T, repo domain.RatingStreakRepo) {

	//setup
	today, err := time.ParseInLocation("02-01-2006", "30-04-2023", time.Local)
	require.NoError(t, err)
	todayDayPrecision := domain.NewDayPrecisionTime(roundTimeToDBResolution(today))

	ctx := context.Background()
	user1 := "test@user"
	wantRSU1S1 := domain.NewRatingStreakFromDB(todayDayPrecision.PrevDay(), todayDayPrecision)
	wantRSU1S2 := domain.NewRatingStreakFromDB(todayDayPrecision.PrevDay().PrevDay().PrevDay(), todayDayPrecision)

	//test : initially, most recent streak should not exist
	_, _, err = repo.GetMostRecentStreak(ctx, user1)
	require.ErrorIsf(t, err, domain.ErrNotFound, "Expected not found error, got %v", err)

	err = repo.UpdateMostRecentRatingStreak(ctx, user1, func(current domain.RatingStreak) (*domain.RatingStreak, error) {
		require.Fail(t, "updateFN should not be called if no streak data exists")
		return nil, fmt.Errorf("test error")
	})
	require.ErrorIsf(t, err, domain.ErrNotFound, "Expected not found error, got %v", err)

	_, _, err = repo.GetLongestStreak(ctx, user1)
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, _, err = repo.GetLongestIndividualStreak(ctx)
	require.ErrorIs(t, err, domain.ErrNotFound)

	//test : create rating streak
	rs1ID, err := repo.CreateRatingStreak(ctx, user1, wantRSU1S1)
	require.NoError(t, err)

	//test : no update on error
	err = repo.UpdateMostRecentRatingStreak(ctx, user1, func(current domain.RatingStreak) (*domain.RatingStreak, error) {
		return nil, fmt.Errorf("you shall not update")
	})
	require.Error(t, err)

	//test :  no update when returning nil
	err = repo.UpdateMostRecentRatingStreak(ctx, user1, func(current domain.RatingStreak) (*domain.RatingStreak, error) {
		return nil, nil
	})
	require.NoError(t, err)

	//test : get rs1 and compare values
	gotRS1, gotRS1ID, err := repo.GetMostRecentStreak(ctx, user1)
	require.NoError(t, err)
	require.Equal(t, rs1ID, gotRS1ID)
	require.Equal(t, wantRSU1S1, gotRS1)

	//test : update rs1 and check updated value
	rs1UpdatedBegin := wantRSU1S1.Begin.PrevDay()
	wantRs1Updated := domain.NewRatingStreakFromDB(rs1UpdatedBegin, wantRSU1S1.End)
	err = repo.UpdateMostRecentRatingStreak(ctx, user1, func(current domain.RatingStreak) (*domain.RatingStreak, error) {
		require.Equal(t, wantRSU1S1, current)

		current.Begin = rs1UpdatedBegin

		return &current, nil
	})
	require.NoError(t, err)
	gotRS1, gotRS1ID, err = repo.GetMostRecentStreak(ctx, user1)
	require.NoError(t, err)
	require.Equal(t, rs1ID, gotRS1ID)
	require.Equal(t, wantRs1Updated, gotRS1)

	//create longer rating streak for user1 and check get longest
	wantRSU1S2ID, err := repo.CreateRatingStreak(ctx, user1, wantRSU1S2)
	require.NoError(t, err)

	gotRS, gotRSID, err := repo.GetLongestStreak(ctx, user1)
	require.NoError(t, err, domain.ErrNotFound)
	require.Equal(t, wantRSU1S2ID, gotRSID)
	require.Equal(t, wantRSU1S2, gotRS)

	//TODO: we need to actually create the users in the users table for this test to work
	/*//create rating streak of equal length for user2, check that GetLongestIndividualStreak returns both users
	_, err = repo.CreateRatingStreak(ctx, user2, wantRS2)
	require.NoError(t, err)

	maxStreakUsers, maxStreak, err := repo.GetLongestIndividualStreak(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{user1, user2}, maxStreakUsers)
	require.Equal(t, wantRS2, maxStreak)*/
}
