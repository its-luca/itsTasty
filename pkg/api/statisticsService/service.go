package statisticsService

import (
	"context"
	"fmt"
	"itsTasty/pkg/api/domain"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/sourcegraph/conc/pool"
)

type TimeSource interface {
	//Now returns the current local time.
	Now() time.Time
}

const AllUsersStreakName = "allUsers"

type StreakService interface {
	UpdateRatingStreaks(ctx context.Context) error
	GetMostRecentAllUsersGroupStreak(ctx context.Context, onlyOngoing bool) (*domain.RatingStreak, error)
	GetMostRecentUserStreaks(ctx context.Context, onlyOngoing bool) ([]UserWithStreak, error)
	GetLongestStreaks(ctx context.Context) (individualUsers []UserWithStreak, allUsersGroup *domain.RatingStreak, err error)
}

type DefaultStreakService struct {
	statsRepo          domain.StatisticsRepo
	vacationStreakRepo domain.RatingStreakRepo
	vacationClient     domain.VacationDataSource
	publicHolidays     domain.PublicHolidayDataSource
	timeSource         TimeSource
}

func NewDefaultStreakService(statsRepo domain.StatisticsRepo, vacationStreakRepo domain.RatingStreakRepo,
	vacationClient domain.VacationDataSource, holidayClient domain.PublicHolidayDataSource, timeSource TimeSource) *DefaultStreakService {
	return &DefaultStreakService{
		statsRepo:          statsRepo,
		vacationStreakRepo: vacationStreakRepo,
		vacationClient:     vacationClient,
		publicHolidays:     holidayClient,
		timeSource:         timeSource,
	}
}

type vacationStreakData struct {
	users              []domain.User
	ratingsToday       []domain.DishRating
	usersOnVacation    domain.UsersOnVacation
	isHolidayOrWeekend bool
}

// fetchData is a helper functions querying the data required in UpdateRatingStreaks.
// The slices/maps users, ratingsToday and usersOnVacation may be empty
func (d *DefaultStreakService) fetchData(ctx context.Context) (vacationStreakData, error) {
	poolCtx, poolCancel := context.WithCancel(ctx)
	defer poolCancel()

	result := vacationStreakData{}
	reqPool := pool.New().WithContext(poolCtx).WithCancelOnError()
	reqPool.Go(func(ctx context.Context) error {
		var err error
		result.users, err = d.statsRepo.GetAllUsers(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				result.users = make([]domain.User, 0)
			} else {
				return fmt.Errorf("failed to fetch users : %w", err)
			}
		}
		return nil
	})

	reqPool.Go(func(ctx context.Context) error {
		var err error
		result.ratingsToday, err = d.statsRepo.GetAllRatingsForDate(ctx, domain.NewDayPrecisionTime(d.timeSource.Now()))
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				result.ratingsToday = make([]domain.DishRating, 0)
			} else {
				return fmt.Errorf("failed to fetch today's ratings : %w", err)
			}
		}
		return nil
	})

	reqPool.Go(func(ctx context.Context) error {
		var err error
		result.usersOnVacation, err = d.vacationClient.Vacations(ctx, domain.NewDayPrecisionTime(d.timeSource.Now()))
		if err != nil {
			return fmt.Errorf("failed to get today's vacations : %w", err)
		}
		return nil
	})

	reqPool.Go(func(ctx context.Context) error {
		var err error
		var isHoliday bool
		date := d.timeSource.Now()
		isHoliday, err = d.publicHolidays.IsPublicHoliday(ctx, date)
		if err != nil {
			return fmt.Errorf("failed to check if %v is public holiday :%w", date, err)
		}
		result.isHolidayOrWeekend = isHoliday || (date.Weekday() == time.Saturday && date.Weekday() == time.Sunday)

		return nil
	})

	poolErr := reqPool.Wait()
	if poolErr != nil {
		return vacationStreakData{}, fmt.Errorf("request in pool failed : %w", poolErr)
	}

	return result, nil
}

func (d *DefaultStreakService) updateRatingStreak(ctx context.Context, streakData vacationStreakData, streakName string, streakUserGroup map[string]interface{}) error {

	isHolidayOrWeekendMap := map[domain.DayPrecisionTime]bool{domain.NewDayPrecisionTime(d.timeSource.Now()): streakData.isHolidayOrWeekend}
	newStreak, err := domain.NewRatingStreak(
		domain.NewDayPrecisionTime(d.timeSource.Now()),
		streakData.ratingsToday,
		streakData.usersOnVacation,
		isHolidayOrWeekendMap, streakUserGroup)
	if err != nil {
		if errors.Is(err, domain.ErrNoStreak) {
			return nil
		}
		return fmt.Errorf("failed to calculate streak for \"%s\": %v", streakName, err)
	}

	createNewStreak := false
	err = d.vacationStreakRepo.UpdateMostRecentRatingStreak(ctx, streakName, func(current domain.RatingStreak) (*domain.RatingStreak, error) {

		//if newStreak is equal to current streak do nothing
		if current.Begin.Equal(newStreak.Begin.Time) && current.End.Equal(current.End.Time) {
			return nil, nil
		}

		//cannot extend prev streak due to gap (assumption, this routine is called daily)
		if !current.End.Equal(domain.NewDayPrecisionTime(d.timeSource.Now()).PrevDay().Time) {
			createNewStreak = true
			return nil, nil
		}

		//have prev streak, that extends to the preceding day -> check if it extends to today

		if current.End.Equal(newStreak.Begin.PrevDay().Time) {
			current.End = newStreak.End
		}

		return &current, nil

	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			createNewStreak = true
		} else {
			return fmt.Errorf("failed to update rating streak for \"%v\" : %w", streakName, err)
		}
	}

	if createNewStreak {
		if _, err := d.vacationStreakRepo.CreateRatingStreak(ctx, streakName, newStreak); err != nil {
			return fmt.Errorf("failed to create new streak for \"%v\" : %w", streakName, err)
		}
	}

	return nil
}

// UpdateRatingStreaks updates the rating streaks for all users currently in the system as well as for the special
// AllUsersStreakName streak. This function must be called daily to function properly. Otherwise, we would have to store
// vacation data of users which is a privacy concern
func (d *DefaultStreakService) UpdateRatingStreaks(ctx context.Context) error {

	//fetch data
	streakData, err := d.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch data :%w", err)
	}

	//update streaks for each user
	for _, user := range streakData.users {
		if err := d.updateRatingStreak(ctx, streakData, user.Email, map[string]interface{}{user.Email: nil}); err != nil {
			return err
		}
	}

	//update special "all users" rating streak
	allUsersMap := map[string]interface{}{}
	for _, v := range streakData.users {
		allUsersMap[v.Email] = nil
	}

	if err := d.updateRatingStreak(ctx, streakData, AllUsersStreakName, allUsersMap); err != nil {
		return err
	}

	return nil

}

// GetMostRecentAllUsersGroupStreak returns the most recent streak for the special "all users" group. If onlyOngoing
// is set, the streak must be unbroken.
// Marker errors: domain.ErrNotFound
func (d *DefaultStreakService) GetMostRecentAllUsersGroupStreak(ctx context.Context, onlyOngoing bool) (*domain.RatingStreak, error) {
	streak, _, err := d.vacationStreakRepo.GetMostRecentStreak(ctx, AllUsersStreakName)
	if err != nil {
		return nil, fmt.Errorf("failed to get most recent streak : %w", err)
	}

	if onlyOngoing && streak.End.Before(domain.NewDayPrecisionTime(d.timeSource.Now()).Time) {
		return nil, domain.ErrNotFound
	}

	return &streak, nil
}

type UserWithStreak struct {
	User   domain.User
	Streak domain.RatingStreak
}

// GetMostRecentUserStreaks returns the most recent streak for each user, if they have one. If onlyOngoing is set,
// only streaks that are currently unbroken are considered. The returned slice
// may be empty.
func (d *DefaultStreakService) GetMostRecentUserStreaks(ctx context.Context, onlyOngoing bool) ([]UserWithStreak, error) {
	users, err := d.statsRepo.GetAllUsers(ctx)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("failed to get users : %w", err)
		}
	}

	result := make([]UserWithStreak, 0)
	for _, user := range users {
		streak, _, err := d.vacationStreakRepo.GetMostRecentStreak(ctx, user.Email)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("failed to get most recent streak for user %v : %w", user.Email, err)
		}

		if onlyOngoing && streak.End.Before(domain.NewDayPrecisionTime(d.timeSource.Now()).Time) {
			continue
		}

		result = append(result, UserWithStreak{
			User:   user,
			Streak: streak,
		})
	}

	return result, nil
}

func (d *DefaultStreakService) GetLongestStreaks(ctx context.Context) (individualUsers []UserWithStreak, allUsersGroup *domain.RatingStreak, err error) {

	fetchPool := pool.New().WithContext(ctx).WithCancelOnError()

	//fetch AllUsersStreakName data
	fetchPool.Go(func(ctx context.Context) error {
		s, _, err := d.vacationStreakRepo.GetLongestStreak(ctx, AllUsersStreakName)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				allUsersGroup = nil
				return nil
			}
			return fmt.Errorf("failed to fetch longest streak for %s : %w", AllUsersStreakName, err)
		}
		allUsersGroup = &s

		return nil
	})

	//fetch individual user streak data
	fetchPool.Go(func(ctx context.Context) error {
		users, streak, err := d.vacationStreakRepo.GetLongestIndividualStreak(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				individualUsers = make([]UserWithStreak, 0)
				return nil
			}
			return fmt.Errorf("failed to fetch longest individual streak : %w", err)
		}
		individualUsers = make([]UserWithStreak, len(users))
		for i, v := range users {
			individualUsers[i] = UserWithStreak{
				User:   domain.User{Email: v},
				Streak: streak,
			}
		}
		return nil
	})

	if poolErr := fetchPool.Wait(); poolErr != nil {
		err = fmt.Errorf("failed to fetch data : %v", poolErr)
	}

	return

}
