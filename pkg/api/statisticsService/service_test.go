package statisticsService

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"itsTasty/pkg/api/adapters/publicHoliday"
	"itsTasty/pkg/api/adapters/vacation"
	"itsTasty/pkg/api/domain"
	"testing"
	"time"
)

//
// Mocks
//

type mockStatsRepo struct {
	users         []domain.User
	ratingsByDate map[domain.DayPrecisionTime][]domain.DishRating
}

func (m mockStatsRepo) GetAllRatingsForDate(_ context.Context, date domain.DayPrecisionTime) ([]domain.DishRating, error) {
	data, ok := m.ratingsByDate[date]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return data, nil
}

func (m mockStatsRepo) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return m.users, nil
}

type mockRatingStreakRepo struct {
	streaks map[string][]domain.RatingStreak
}

func (m mockRatingStreakRepo) GetLongestIndividualStreak(ctx context.Context) ([]string, domain.RatingStreak, error) {
	panic("implemente me")
}
func (m mockRatingStreakRepo) GetLongestStreak(ctx context.Context, name string) (domain.RatingStreak, int, error) {
	panic("implement me")
}
func (m mockRatingStreakRepo) UpdateMostRecentRatingStreak(ctx context.Context, name string, updateFN domain.StreakUpdateFN) error {
	current, _, err := m.GetMostRecentStreak(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get most recent streak: %w", err)
	}

	updated, err := updateFN(current)
	if err != nil {
		return fmt.Errorf("updateFN failed : %v", err)
	}
	if updated == nil {
		return nil
	}

	s := m.streaks[name]
	s[len(s)-1] = *updated
	return nil
}

func (m mockRatingStreakRepo) CreateRatingStreak(ctx context.Context, name string, data domain.RatingStreak) (int, error) {
	if _, ok := m.streaks[name]; !ok {
		m.streaks[name] = make([]domain.RatingStreak, 0)
	}
	m.streaks[name] = append(m.streaks[name], data)
	return 1, nil
}

func (m mockRatingStreakRepo) GetMostRecentStreak(ctx context.Context, name string) (domain.RatingStreak, int, error) {
	s, ok := m.streaks[name]
	if !ok {
		return domain.RatingStreak{}, 0, domain.ErrNotFound
	}
	return s[len(s)-1], 1, nil
}

type mockTimeSource struct {
	CurrentTime time.Time
}

func (m *mockTimeSource) Now() time.Time {
	return m.CurrentTime
}

func NewMockTimeSourceToday() *mockTimeSource {
	return &mockTimeSource{CurrentTime: time.Now()}
}

func (m *mockTimeSource) AdvanceBy(d time.Duration) {
	m.CurrentTime = m.CurrentTime.Add(d)
}

// MustNewMockTimeSource. date format "DD-MM-YYYY"
func mustNewMockTimeSource(startDate string) *mockTimeSource {
	t, err := time.ParseInLocation("02-01-2006", startDate, time.Local)
	if err != nil {
		panic(fmt.Sprintf("failed to parse timestring %v : %v", startDate, err))
	}
	return &mockTimeSource{CurrentTime: t}
}

//
// Test Functions
//

func TestDefaultStreakService_GetMostRecentUserStreaks_NoUsers(t *testing.T) {

	//
	//setup env
	//

	statsRepo := mockStatsRepo{
		users:         make([]domain.User, 0),
		ratingsByDate: make(map[domain.DayPrecisionTime][]domain.DishRating),
	}

	streakRepo := mockRatingStreakRepo{streaks: make(map[string][]domain.RatingStreak)}

	vacationClint := vacation.NewEmptyVacationClient()

	holidayClient, err := publicHoliday.NewDefaultRegionHolidayChecker("Schleswig-Holstein")
	require.NoError(t, err)

	timeSource := NewMockTimeSourceToday()

	//
	//test
	//

	service := NewDefaultStreakService(statsRepo, streakRepo, vacationClint, holidayClient, timeSource)

	ctx := context.Background()
	err = service.UpdateRatingStreaks(ctx)
	require.NoError(t, err)

	gotUserStreaks, err := service.GetMostRecentUserStreaks(ctx, true)
	require.NoError(t, err)
	require.Empty(t, gotUserStreaks)

	_, err = service.GetMostRecentAllUsersGroupStreak(ctx, true)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDefaultStreakService_GetMostRecentUserStreaks_NoPrevStreak(t *testing.T) {

	//
	//setup env
	//
	timeSource := mustNewMockTimeSource("01-02-2023")

	user1 := domain.User{Email: "user1@test.user"}
	user2 := domain.User{Email: "user2@test.user"}
	day1 := domain.NewDayPrecisionTime(timeSource.Now())

	user1Rating1 := domain.DishRating{
		Who:        user1.Email,
		Value:      domain.ThreeStars,
		RatingWhen: timeSource.Now(),
	}

	statsRepo := mockStatsRepo{
		users:         []domain.User{user1, user2},
		ratingsByDate: map[domain.DayPrecisionTime][]domain.DishRating{day1: {user1Rating1}},
	}

	streakRepo := mockRatingStreakRepo{streaks: make(map[string][]domain.RatingStreak)}

	vacationClient := vacation.NewEmptyVacationClient()

	holidayClient, err := publicHoliday.NewDefaultRegionHolidayChecker("Schleswig-Holstein")
	require.NoError(t, err)

	//
	//test
	//

	service := NewDefaultStreakService(statsRepo, streakRepo, vacationClient, holidayClient, timeSource)

	ctx := context.Background()
	err = service.UpdateRatingStreaks(ctx)
	require.NoError(t, err)

	gotUserStreaks, err := service.GetMostRecentUserStreaks(ctx, true)
	wantUserStreaks := []UserWithStreak{
		{
			User: user1,
			Streak: domain.RatingStreak{
				Begin: day1,
				End:   day1,
			},
		},
	}
	require.NoError(t, err)
	require.ElementsMatch(t, wantUserStreaks, gotUserStreaks)

	allUsersStreak, err := service.GetMostRecentAllUsersGroupStreak(ctx, true)
	wantAllUsersStreak := domain.RatingStreak{
		Begin: day1,
		End:   day1,
	}
	require.NoError(t, err)
	require.Equal(t, wantAllUsersStreak, *allUsersStreak)
}

func TestDefaultStreakService_GetMostRecentUserStreaks_PrevStreak_Can_Extend(t *testing.T) {

	//
	//setup env
	//
	timeSource := mustNewMockTimeSource("01-02-2023")

	user1 := domain.User{Email: "user1@test.user"}
	user2 := domain.User{Email: "user2@test.user"}
	yesterday := domain.NewDayPrecisionTime(timeSource.Now()).PrevDay()
	today := yesterday.NextDay()

	user1Rating1 := domain.DishRating{
		Who:        user1.Email,
		Value:      domain.ThreeStars,
		RatingWhen: timeSource.Now(),
	}

	statsRepo := mockStatsRepo{
		users:         []domain.User{user1, user2},
		ratingsByDate: map[domain.DayPrecisionTime][]domain.DishRating{today: {user1Rating1}},
	}

	streakRepo := mockRatingStreakRepo{streaks: map[string][]domain.RatingStreak{
		user1.Email: {
			{
				Begin: yesterday,
				End:   yesterday,
			},
		},
		AllUsersStreakName: {
			{
				Begin: yesterday,
				End:   yesterday,
			},
		},
	}}

	vacationClient := vacation.NewEmptyVacationClient()

	holidayClient, err := publicHoliday.NewDefaultRegionHolidayChecker("Schleswig-Holstein")
	require.NoError(t, err)

	//
	//test
	//

	service := NewDefaultStreakService(statsRepo, streakRepo, vacationClient, holidayClient, timeSource)

	ctx := context.Background()
	err = service.UpdateRatingStreaks(ctx)
	require.NoError(t, err)

	gotUserStreaks, err := service.GetMostRecentUserStreaks(ctx, true)
	wantUserStreaks := []UserWithStreak{
		{
			User: user1,
			Streak: domain.RatingStreak{
				Begin: yesterday,
				End:   today,
			},
		},
	}
	require.NoError(t, err)
	require.ElementsMatch(t, wantUserStreaks, gotUserStreaks)

	allUsersStreak, err := service.GetMostRecentAllUsersGroupStreak(ctx, true)
	wantAllUsersStreak := domain.RatingStreak{
		Begin: yesterday,
		End:   today,
	}
	require.NoError(t, err)
	require.Equal(t, wantAllUsersStreak, *allUsersStreak)
}

func TestDefaultStreakService_GetMostRecentUserStreaks_PrevStreak_Cannot_Extend(t *testing.T) {

	//
	//setup env
	//
	timeSource := mustNewMockTimeSource("01-02-2023")

	user1 := domain.User{Email: "user1@test.user"}
	user2 := domain.User{Email: "user2@test.user"}
	dayBeforeYesterday := domain.NewDayPrecisionTime(timeSource.Now()).PrevDay().PrevDay()
	yesterday := dayBeforeYesterday.NextDay()
	today := yesterday.NextDay()

	user1Rating1 := domain.DishRating{
		Who:        user1.Email,
		Value:      domain.ThreeStars,
		RatingWhen: timeSource.Now(),
	}

	statsRepo := mockStatsRepo{
		users:         []domain.User{user1, user2},
		ratingsByDate: map[domain.DayPrecisionTime][]domain.DishRating{today: {user1Rating1}},
	}

	streakRepo := mockRatingStreakRepo{streaks: map[string][]domain.RatingStreak{
		user1.Email: {
			{
				Begin: dayBeforeYesterday,
				End:   dayBeforeYesterday,
			},
		},
		AllUsersStreakName: {
			{
				Begin: dayBeforeYesterday,
				End:   dayBeforeYesterday,
			},
		},
	}}

	vacationClient := vacation.NewEmptyVacationClient()

	holidayClient, err := publicHoliday.NewDefaultRegionHolidayChecker("Schleswig-Holstein")
	require.NoError(t, err)

	//
	//test
	//

	service := NewDefaultStreakService(statsRepo, streakRepo, vacationClient, holidayClient, timeSource)

	ctx := context.Background()
	err = service.UpdateRatingStreaks(ctx)
	require.NoError(t, err)

	gotUserStreaks, err := service.GetMostRecentUserStreaks(ctx, true)
	wantUserStreaks := []UserWithStreak{
		{
			User: user1,
			Streak: domain.RatingStreak{
				Begin: today,
				End:   today,
			},
		},
	}
	require.NoError(t, err)
	require.ElementsMatch(t, wantUserStreaks, gotUserStreaks)

	allUsersStreak, err := service.GetMostRecentAllUsersGroupStreak(ctx, true)
	wantAllUsersStreak := domain.RatingStreak{
		Begin: today,
		End:   today,
	}
	require.NoError(t, err)
	require.Equal(t, wantAllUsersStreak, *allUsersStreak)

	//advance time by one day -> check that onlyOngoing=true returns no more streaks while onlyOngoing=false returns
	//the same result

	timeSource.AdvanceBy(24 * time.Hour)
	gotUserStreaks, err = service.GetMostRecentUserStreaks(ctx, true)
	require.NoError(t, err)
	require.ElementsMatch(t, []UserWithStreak{}, gotUserStreaks)

	allUsersStreak, err = service.GetMostRecentAllUsersGroupStreak(ctx, true)
	require.ErrorIs(t, err, domain.ErrNotFound)

	gotUserStreaks, err = service.GetMostRecentUserStreaks(ctx, false)
	require.NoError(t, err)
	require.ElementsMatch(t, wantUserStreaks, gotUserStreaks)

	allUsersStreak, err = service.GetMostRecentAllUsersGroupStreak(ctx, false)
	require.NoError(t, err)
	require.Equal(t, wantAllUsersStreak, *allUsersStreak)
}
