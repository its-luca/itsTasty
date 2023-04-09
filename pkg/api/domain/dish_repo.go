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

	//GetOrCreateDish fetches the dish or creates a new one if it does not exist. The same is true
	//for the location denoted by servedAt.
	//The two bool results indicate whether a new dish and/or a new location was created.
	//The int64 result is the id of the dish
	GetOrCreateDish(ctx context.Context, dishName string, servedAt string) (*Dish, bool, bool, int64, error)
	//GetDishByName fetches the dish. The second result is the id of the dish
	GetDishByName(ctx context.Context, dishName, servedAt string) (dish *Dish, dishID int64, err error)
	//GetDishByID fetches the dish. Returns domain.ErrNotFound if dish could not be found
	GetDishByID(ctx context.Context, dishID int64) (dish *Dish, err error)

	//GetDishByDate returns dishIDs for all dishes served at when optionally restricted to those served by the given location
	GetDishByDate(ctx context.Context, when time.Time, optionalLocation *string) ([]int64, error)

	//UpdateMostRecentServing calls updateFN with the most recent serving for dishID (which may be nil)
	//and adds a new serving if the function returns a non nil time value
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
	GetMergedDishByID(ctx context.Context, id int64) (*MergedDish, error)

	GetMostRecentDishForMergedDish(ctx context.Context, mergedDishID int64) (*Dish, int64, error)

	//DeleteMergedDish removes all dish from the merged dish and deletes the merged dish entry
	//(but not the individual dishes)
	DeleteMergedDish(ctx context.Context, mergedDishName, servedAt string) error
	DeleteMergedDishByID(ctx context.Context, mergedDishID int64) error

	//UpdateMergedDishByID calls updateFN with the current value of the merged dish. If updateFN does not return an error
	//the db entry is updated with the returned value
	//Fails with domain.ErrNotFound if id is not found
	UpdateMergedDishByID(ctx context.Context, id int64, updateFN func(current *MergedDish) (*MergedDish, error)) (err error)

	//
	// Manipulate DishRating struct
	//

	//GetRatings returns all ratings of the user for the dish, unless onlyMostRecent is true in which case only
	//the most recent rating is returned. Ratings are sorted in descending order, i.e. the most recent rating is at index
	//0. Second result is the id of the rating
	//Returns domain.ErrNotFound if there are no ratings yet
	GetRatings(ctx context.Context, userEmail string, dishID int64, onlyMostRecent bool) ([]DishRating, error)

	//CreateOrUpdateRating calls updateFN with the most recent rating (or nil if no rating exists)
	//If updateFN returns nil for rating, no changes are made. Otherwise, the returned value either replaces the
	//most recent rating or is added as a new rating, depending on the value of createNew
	CreateOrUpdateRating(ctx context.Context, userEmail string, dishID int64,
		updateFN func(currentRating *DishRating) (updatedRating *DishRating, createNew bool, err error)) (err error)
	//GetAllRatingsForDish returns all ratings for the dish. If a dish has multiple servings this means that
	//there may be up to one rating per user per serving
	GetAllRatingsForDish(ctx context.Context, dishID int64) ([]DishRating, error)

	//DropRepo drops all tables related to this repo
	DropRepo(ctx context.Context) error

	//Close closes the db connection
	Close() error

	//IsDishPartOfMergedDish if dish is part of merged dish, (true, <merged dish id>,nil) is returned. If not part
	//of merged dish (false, <invalid> ,nil)
	IsDishPartOfMergedDish(ctx context.Context, dishName string, servedAt string) (bool, int64, error)
	IsDishPartOfMergedDisByID(ctx context.Context, dishID int64) (bool, int64, error)
}
