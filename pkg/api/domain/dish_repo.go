package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type DishRepo interface {
	GetOrCreateDish(ctx context.Context, dishName string) (*Dish, bool, error)
	GetDish(ctx context.Context, dishName string) (*Dish, error)
	UpdateDish(ctx context.Context, dishName string, updateFn func(d *Dish) error) error
	Close(ctx context.Context) error
}
