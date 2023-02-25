package domain

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrDishAlreadyMerged = errors.New("dish already part of merged dish")

type DishRepo interface {

	//
	// Manipulate Dish struct
	//

	//GetOrCreateDish fetches the dish or creates a new one if it does not exist. The same is true
	//for the location denoted by servedAt.
	//The two bool results indicate whether a new dish and/or a new location was created.
	//The int64 result is the id of the dish
	//TODO: update for merged dishes
	GetOrCreateDish(ctx context.Context, dishName string, servedAt string) (*Dish, bool, bool, int64, error)
	//GetDishByName fetches the dish. The second result is the id of the dish
	//TODO: update for merged dishes
	GetDishByName(ctx context.Context, dishName, servedAt string) (dish *Dish, dishID int64, err error)
	//GetDishByID fetches the dish
	//TODO: update for merged dishes
	GetDishByID(ctx context.Context, dishID int64) (dish *Dish, err error)

	//GetDishByDate returns dishIDs for all dishes served at when optionally restricted to those served by the given location
	//TODO: update for merged dishes
	GetDishByDate(ctx context.Context, when time.Time, optionalLocation *string) ([]int64, error)

	//UpdateMostRecentServing calls updateFN with the most recent serving for dishID (which may be nil)
	//and adds a new serving if the function returns a non nil time value
	//TODO: update for merged dishes
	UpdateMostRecentServing(ctx context.Context, dishID int64,
		updateFN func(currenMostRecent *time.Time) (*time.Time, error)) (err error)

	//GetAllDishIDs returns a slice with all dish ids
	GetAllDishIDs(ctx context.Context) ([]int64, error)

	//CRUD for merged dishes

	//CreateMergedDish creates a new merged dish with name mergedDishName that consists of/merges dish1Name and dish2Name
	// On success the merged dish and its db id are returned.
	// If either dish1Name or dish2Name are already part of a merged dish, error is set to ErrDishAlreadyMerged
	CreateMergedDish(ctx context.Context, dish *MergedDish) (int64, error)

	GetMergedDish(ctx context.Context, name, servedAt string) (*MergedDish, int64, error)

	//AddDishToMergedDish adds the dish dishName to the merged dish mergedDishName
	//If the dish is already part of another merged dish, error is set to ErrDishAlreadyMerged
	AddDishToMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) error
	//RemoveDishFromMergedDish removes the dish from the merged dish
	//If the dish is not part of the merged dish, error is set to ErrNotFound
	RemoveDishFromMergedDish(ctx context.Context, mergedDishName, servedAt, dishName string) error
	//DeleteMergedDish removes all dish from the merged dish and deletes the merged dish entry
	//(but not the individual dishes)
	DeleteMergedDish(ctx context.Context, mergedDishName, servedAt string) error

	//
	// Manipulate DishRating struct
	//

	/*TODO:
	 1) Dishes can have one rating per serving.
		Rating a dish creates a new rating if there has been a new serving
		since the last rating. Otherwise the old rating is updated
		You cannot directly rate a merged dish. Instead changing a merged dish will always update the data of the most
		recently served dish, contained in the merged dish. This makes unmerging easy
	 2) When Merging two dishes, the most recent rating of all merged dishes becomes the one that is evaluated in the star
	    rating. Edge case: dishes served on same day
	*/

	//GetRating returns the rating of the user for the dish. Second result is the id of the rating
	//TODO: update for merged dishes
	GetRating(ctx context.Context, userEmail string, dishID int64) (*DishRating, error)
	//SetOrCreateRating creates or overwrites the rating of the user for the given dish
	//TODO: update for merged dishes
	SetOrCreateRating(ctx context.Context, userEmail string, dishID int64, rating DishRating) (bool, error)
	//GetAllRatingsForDish returns all ratings for the dish
	//TODO: update for merged dishes
	GetAllRatingsForDish(ctx context.Context, dishID int64) (*DishRatings, error)

	//DropRepo drops all tables related to this repo
	DropRepo(ctx context.Context) error

	//Close closes the db connection
	Close() error
}
