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

	GetOrCreateDish(ctx context.Context, dishName string) (*Dish, bool, error)
	GetDishByName(ctx context.Context, dishName string) (*Dish, int, error)
	GetDishByID(ctx context.Context, dishID int) (*Dish, error)
	UpdateDish(ctx context.Context, dishID int, updateFn func(d *Dish) error) error
	GetAllDishIDs(ctx context.Context) ([]int, error)

	//
	// Manipulate DishRating struct
	//

	GetRating(ctx context.Context, userEmail string, dishID int) (*DishRating, error)
	SetRating(ctx context.Context, userEmail string, dishID int, rating DishRating) error
	GetAllRatingsForDish(ctx context.Context, dishID int) (*DishRatings, error)

	Close(ctx context.Context) error
}
