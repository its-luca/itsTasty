package domain

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type DishRepo interface {

	//
	// Manipulate Dish struct
	//

	//GetOrCreateDish fetches the dish or creates a new one if it does not exist
	//Second result indicates if new dish was created. Third result is the id of the new dish
	GetOrCreateDish(ctx context.Context, dishName string) (dish *Dish, isNew bool, dishID int64, err error)
	//GetDishByName fetches the dish. The second result is the id of the dish
	GetDishByName(ctx context.Context, dishName string) (dish *Dish, dishID int64, err error)
	//GetDishByID fetches the dish
	GetDishByID(ctx context.Context, dishID int64) (dish *Dish, err error)

	//UpdateMostRecentServing calls updateFN with the most recent serving for dishID (which may be nil)
	//and adds a new serving if the function returns a non nil time value
	UpdateMostRecentServing(ctx context.Context, dishID int64,
		updateFN func(currenMostRecent *time.Time) (*time.Time, error)) (err error)

	//GetAllDishIDs returns a slice with all dish ids
	GetAllDishIDs(ctx context.Context) ([]int64, error)

	//
	// Manipulate DishRating struct
	//

	//GetRating returns the rating of the user for the dish. Second result is the id of the rating
	GetRating(ctx context.Context, userEmail string, dishID int64) (*DishRating, error)
	//SetOrCreateRating creates or overwrites the rating of the use for the given dish
	SetOrCreateRating(ctx context.Context, userEmail string, dishID int64, rating DishRating) (bool, error)
	//GetAllRatingsForDish returns all ratings for the dish
	GetAllRatingsForDish(ctx context.Context, dishID int64) (*DishRatings, error)

	//DropRepo drops all tables related to this repo
	DropRepo(ctx context.Context) error

	//Close closes the db connection
	Close() error
}
