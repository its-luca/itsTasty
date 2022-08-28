package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type DishRepo interface {

	//
	// Manipulate Dish struct
	//

	//GetOrCreateDish fetches the dish or creates a new one if it does not exist
	//Second result indicates if new dish was created. Third result is the id of the new dish
	GetOrCreateDish(ctx context.Context, dishName string) (*Dish, bool, int, error)
	//GetDishByName fetches the dish. The second result is the id of the dish
	GetDishByName(ctx context.Context, dishName string) (*Dish, int, error)
	//GetDishByID fetches the dish
	GetDishByID(ctx context.Context, dishID int) (*Dish, error)
	//UpdateDish updates the dish, if it does exist
	UpdateDish(ctx context.Context, dishID int, updateFn func(d *Dish) error) error
	//GetAllDishIDs returns a slice with all dish ids
	GetAllDishIDs(ctx context.Context) ([]int, error)

	//
	// Manipulate DishRating struct
	//

	//GetRating returns the rating of the user for the dish. Second result is the id of the rating
	GetRating(ctx context.Context, userEmail string, dishID int) (*DishRating, int, error)
	//SetRating creates or overwrites the rating of the use for the given dish
	SetRating(ctx context.Context, userEmail string, dishID int, rating DishRating) error
	//GetAllRatingsForDish returns all ratings for the dish
	GetAllRatingsForDish(ctx context.Context, dishID int) (*DishRatings, error)

	Close(ctx context.Context) error
}
