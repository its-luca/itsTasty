package domain

import (
	"context"
	"time"
)

type VacationDataSource interface {
	Vacations(ctx context.Context, day DayPrecisionTime) (UsersOnVacation, error)
}

type PublicHolidayDataSource interface {
	IsPublicHoliday(ctx context.Context, date time.Time) (bool, error)
}

type StreakUpdateFN = func(current RatingStreak) (*RatingStreak, error)
type RatingStreakRepo interface {
	//UpdateMostRecentRatingStreak calls updateFN with the most recent streak of the given user/user group.
	//If updateFN returns (nil,nil) nothing is updated. If updateFN does not return an error, the  return value is used to update
	//the current streak.
	//Marker errors: returns domain.ErrNotFound if no streak exists
	UpdateMostRecentRatingStreak(ctx context.Context, name string, updateFN StreakUpdateFN) error
	//CreateRatingStreak creates a new rating streaks. For a given name, there may be only one streak for each
	//(Start Date,End Date) tuple
	CreateRatingStreak(ctx context.Context, name string, data RatingStreak) (int, error)
	//GetMostRecentStreak returns the most recent streak entry for the given user/user group name.
	//Marker errors ErrNotFound
	GetMostRecentStreak(ctx context.Context, name string) (RatingStreak, int, error)
	//GetLongestStreak returns the longest streak for the given user/user group name.
	//Marker errors ErrNotFound
	GetLongestStreak(ctx context.Context, name string) (RatingStreak, int, error)
	//GetLongestIndividualStreak returns the longest streak only considering single users
	//Marker errors ErrNotFound
	GetLongestIndividualStreak(ctx context.Context) ([]string, RatingStreak, error)
}

type StatisticsRepo interface {
	GetAllUsers(ctx context.Context) ([]User, error)
	GetAllRatingsForDate(ctx context.Context, date DayPrecisionTime) ([]DishRating, error)
}
