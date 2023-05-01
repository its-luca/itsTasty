package statisticsService

import (
	"context"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/sourcegraph/conc/pool"
	"itsTasty/pkg/api/domain"
	"time"
)

type TimeSource interface {
	//Now returns the current local time.
	Now() time.Time
}

const AllUsersStreakName = "allUsers"

type StreakService interface {
	UpdateRatingStreaks(ctx context.Context) error
	GetMostRecentAllUsersGroupStreak(ctx context.Context) (*domain.RatingStreak, error)
	GetMostRecentUserStreaks(ctx context.Context) ([]UserWithStreak, error)
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

// fetchData is a helper functions querying the data required in UpdateRatingStreaks.
// The slices/maps users, ratingsToday and usersOnVacation may be empty
func (d *DefaultStreakService) fetchData(ctx context.Context) (users []domain.User, ratingsToday []domain.DishRating,
	usersOnVacation domain.UsersOnVacation, isHolidayOrWeekend bool, err error) {
	poolCtx, poolCancel := context.WithCancel(ctx)
	defer poolCancel()

	reqPool := pool.New().WithContext(poolCtx).WithCancelOnError()

	reqPool.Go(func(ctx context.Context) error {
		var err error
		users, err = d.statsRepo.GetAllUsers(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				users = make([]domain.User, 0)
			} else {
				return fmt.Errorf("failed to fetch users : %w", err)
			}
		}
		return nil
	})

	reqPool.Go(func(ctx context.Context) error {
		var err error
		ratingsToday, err = d.statsRepo.GetAllRatingsForDate(ctx, domain.NewDayPrecisionTime(d.timeSource.Now()))
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				ratingsToday = make([]domain.DishRating, 0)
			} else {
				return fmt.Errorf("failed to fetch today's ratings : %w", err)
			}
		}
		return nil
	})

	reqPool.Go(func(ctx context.Context) error {
		var err error
		usersOnVacation, err = d.vacationClient.Vacations(ctx, domain.NewDayPrecisionTime(d.timeSource.Now()))
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
		isHolidayOrWeekend = isHoliday || (date.Weekday() == time.Saturday && date.Weekday() == time.Sunday)

		return nil
	})

	poolErr := reqPool.Wait()
	if poolErr != nil {
		err = fmt.Errorf("request in pool failed : %w", poolErr)
	}

	return
}

// UpdateRatingStreaks updates the rating streaks for all users currently in the system as well as for the special
// AllUsersStreakName streak. This function must be called daily to function properly. Otherwise, we would have to store
// vacation data of users which is a privacy concern
func (d *DefaultStreakService) UpdateRatingStreaks(ctx context.Context) error {

	//fetch data
	users, ratingsToday, usersOnVacation, isHolidayOrWeekend, err := d.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch data :%w", err)
	}
	isHolidayOrWeekendMap := map[domain.DayPrecisionTime]bool{domain.NewDayPrecisionTime(d.timeSource.Now()): isHolidayOrWeekend}

	//update streaks for each user
	for _, user := range users {
		newStreak, err := domain.NewRatingStreak(domain.NewDayPrecisionTime(d.timeSource.Now()), ratingsToday, usersOnVacation,
			isHolidayOrWeekendMap, map[string]interface{}{user.Email: nil})
		if err != nil {
			if errors.Is(err, domain.ErrNoStreak) {
				continue
			}
			return fmt.Errorf("failed to determine streak for user %v : %v", user, err)
		}

		//check if today's streak extends previous streak
		cannotExtendPrevStreak := false
		err = d.vacationStreakRepo.UpdateMostRecentRatingStreak(ctx, user.Email, func(current domain.RatingStreak) (*domain.RatingStreak, error) {

			//cannot extend prev streak due to gap (assumption, this routine is called daily)
			yesterday := domain.NewDayPrecisionTime(d.timeSource.Now()).PrevDay()
			if !current.End.Equal(yesterday.Time) {
				cannotExtendPrevStreak = true
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
				cannotExtendPrevStreak = true
			} else {
				return fmt.Errorf("failed to update rating streak for user %v : %w", user, err)
			}
		}

		if cannotExtendPrevStreak {
			if _, err := d.vacationStreakRepo.CreateRatingStreak(ctx, user.Email, newStreak); err != nil {
				return fmt.Errorf("failed to create new streak for user %v : %w", user, err)
			}
		}

	}

	//update special "all users" rating streak
	//TODO: try to unify code with the update code for individual users
	allUsersMap := map[string]interface{}{}
	for _, v := range users {
		allUsersMap[v.Email] = nil
	}

	newStreak, err := domain.NewRatingStreak(domain.NewDayPrecisionTime(d.timeSource.Now()), ratingsToday, usersOnVacation,
		isHolidayOrWeekendMap, allUsersMap)
	if err != nil {
		if errors.Is(err, domain.ErrNoStreak) {
			return nil
		}
		return fmt.Errorf("failed to calculate \"all users\" streak : %v", err)
	}

	//check if today's streak extends previous streak
	cannotExtendPrevStreak := false
	err = d.vacationStreakRepo.UpdateMostRecentRatingStreak(ctx, AllUsersStreakName, func(current domain.RatingStreak) (*domain.RatingStreak, error) {

		//cannot extend prev streak due to gap (assumption, this routine is called daily)
		if !current.End.Equal(domain.NewDayPrecisionTime(d.timeSource.Now()).PrevDay().Time) {
			cannotExtendPrevStreak = true
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
			cannotExtendPrevStreak = true
		} else {
			return fmt.Errorf("failed to update rating streak for all users group : %w", err)
		}
	}

	if cannotExtendPrevStreak {
		if _, err := d.vacationStreakRepo.CreateRatingStreak(ctx, AllUsersStreakName, newStreak); err != nil {
			return fmt.Errorf("failed to create new streak for all users group : %w", err)
		}
	}

	return nil

}

// GetMostRecentAllUsersGroupStreak returns the most recent streak for the special "all users" group
// Marker errors: domain.ErrNotFound
func (d *DefaultStreakService) GetMostRecentAllUsersGroupStreak(ctx context.Context) (*domain.RatingStreak, error) {
	streak, _, err := d.vacationStreakRepo.GetMostRecentStreak(ctx, AllUsersStreakName)
	if err != nil {
		return nil, fmt.Errorf("failed to get most recent streak : %w", err)
	}
	return &streak, nil
}

type UserWithStreak struct {
	User             domain.User
	MostRecentStreak domain.RatingStreak
}

// GetMostRecentUserStreaks returns the most recent streak for each user, if they have one. The returned slice
// may be empty if not user has any streaks so far
func (d *DefaultStreakService) GetMostRecentUserStreaks(ctx context.Context) ([]UserWithStreak, error) {
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

		result = append(result, UserWithStreak{
			User:             user,
			MostRecentStreak: streak,
		})
	}

	return result, nil
}
