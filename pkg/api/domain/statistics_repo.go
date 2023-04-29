package domain

import (
	"context"
	"time"
)

type WorkdayRepo interface {
	IsPublicHoliday(ctx context.Context, t time.Time) (bool, error)
	GetVacations(ctx context.Context, t time.Time) (UsersOnVacation, error)
}

type StatisticsRepo interface {

	// GetAllUserRatingsSorted GetAllUserRatings returns all user rating data in the requested sort order skipping
	//the first offset values and returning at most limit many entries.
	//If userEmail is not nil, only votes of the given user are considered
	//The second return value
	// is the total amount of user ratings available
	GetAllUserRatingsSorted(ctx context.Context, sortAsc bool, offset, limit int, userEmail *string) ([]DishRating, int, error)
}
