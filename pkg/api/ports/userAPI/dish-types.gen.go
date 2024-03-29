// Package userAPI provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package userAPI

import (
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
)

// Defines values for GetDishRespRatingOfUser.
const (
	GetDishRespRatingOfUserN1 GetDishRespRatingOfUser = 1
	GetDishRespRatingOfUserN2 GetDishRespRatingOfUser = 2
	GetDishRespRatingOfUserN3 GetDishRespRatingOfUser = 3
	GetDishRespRatingOfUserN4 GetDishRespRatingOfUser = 4
	GetDishRespRatingOfUserN5 GetDishRespRatingOfUser = 5
)

// Defines values for RateDishReqRating.
const (
	RateDishReqRatingN1 RateDishReqRating = 1
	RateDishReqRatingN2 RateDishReqRating = 2
	RateDishReqRatingN3 RateDishReqRating = 3
	RateDishReqRatingN4 RateDishReqRating = 4
	RateDishReqRatingN5 RateDishReqRating = 5
)

// BasicError defines model for BasicError.
type BasicError struct {
	What *string `json:"what,omitempty"`
}

// ContainedDishEntry Information about dish contained in mergeddish
type ContainedDishEntry struct {
	// Id dish ID
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateMergedDishReq Request to create a new MergedDish
type CreateMergedDishReq struct {
	// MergedDishes Array of dish ids that should be merged. All dishes must be served at the same location \ and cannot be part of any other merged dishes.  At least two dishes must be provided
	MergedDishes []int64 `json:"mergedDishes"`

	// Name Name of the merged dish. May be equal to existing dishes but for a given \ location there may not be another merged dish with the same name.
	Name string `json:"name"`
}

// CreateMergedDishResp Success response for MergedDish creation
type CreateMergedDishResp struct {
	// MergedDishID ID of the newly created merged dish
	MergedDishID int64 `json:"mergedDishID"`
}

// GetAllDishesRespEntry Entry in the result array returned by GetAllDishesResponse
type GetAllDishesRespEntry struct {
	// Id dishID
	Id int64 `json:"id"`

	// MergedDishID Optional field, if this dish is part of a merged dish
	MergedDishID *int64 `json:"mergedDishID,omitempty"`

	// Name Name of this dish
	Name string `json:"name"`

	// ServedAt Location where this dish is served at
	ServedAt string `json:"servedAt"`
}

// GetAllDishesResponse defines model for GetAllDishesResponse.
type GetAllDishesResponse struct {
	Data []GetAllDishesRespEntry `json:"data"`
}

// GetDishResp Detailed description of a dish
type GetDishResp struct {
	// AvgRating Average rating for this dish. Omitted if there are no votes yet
	AvgRating *float32 `json:"avgRating,omitempty"`

	// MergedDishID If set, the dish is part of this merged dish
	MergedDishID *int64 `json:"mergedDishID,omitempty"`

	// Name Name of the dish
	Name string `json:"name"`

	// OccurrenceCount Amount of times this dish occurred
	OccurrenceCount int `json:"occurrenceCount"`

	// RatingOfUser Most recent rating for this dish of the requesting user. Omitted if the user has not rated yet.
	RatingOfUser *GetDishRespRatingOfUser `json:"ratingOfUser,omitempty"`

	// Ratings Ratings for this dish. Includes up to one vote per user per serving. Keys mean rating, values mean ratings with that amount of stars. If more than zero votes are present avgRating field contains the average rating.
	Ratings map[string]int `json:"ratings"`

	// RecentOccurrences Most recent occurrences of the dish. Might not contain the whole history
	RecentOccurrences []openapi_types.Date `json:"recentOccurrences"`

	// ServedAt Location where this dish is served
	ServedAt string `json:"servedAt"`
}

// GetDishRespRatingOfUser Most recent rating for this dish of the requesting user. Omitted if the user has not rated yet.
type GetDishRespRatingOfUser int

// GetMergeCandidatesResp defines model for GetMergeCandidatesResp.
type GetMergeCandidatesResp struct {
	Candidates []GetMergeCandidatesRespEntry `json:"candidates"`
}

// GetMergeCandidatesRespEntry defines model for GetMergeCandidatesRespEntry.
type GetMergeCandidatesRespEntry struct {
	// DishID dish ID
	DishID int64 `json:"dishID"`

	// DishName dish name
	DishName string `json:"dishName"`

	// MergedDishID If set, this dish is already part of a merged dish
	MergedDishID *int64 `json:"mergedDishID,omitempty"`
}

// GetUsersMeResp Information about the requesting user
type GetUsersMeResp struct {
	Email string `json:"email"`
}

// MergedDishManagementData Management Data for merged dish
type MergedDishManagementData struct {
	// ContainedDishes Information about contained dishes
	ContainedDishes []ContainedDishEntry `json:"containedDishes"`

	// Name Name of the merged dish
	Name string `json:"name"`

	// ServedAt Location the merged dish is served at
	ServedAt string `json:"servedAt"`
}

// MergedDishUpdateReq Representation of a merged dish
type MergedDishUpdateReq struct {
	// AddDishIDs If present, these IDs are added to the merged dish.
	AddDishIDs *[]int64 `json:"addDishIDs,omitempty"`

	// Name If present, the merged dish will be renamed to this
	Name *string `json:"name,omitempty"`

	// RemoveDishIDs If present, these IDs are removed from the merged dish. At least two dish must remain. \ To delete a merge dish, use DELETE instead of PATCH
	RemoveDishIDs *[]int64 `json:"removeDishIDs,omitempty"`
}

// RateDishReq Request to vote for a dish
type RateDishReq struct {
	Rating RateDishReqRating `json:"rating"`
}

// RateDishReqRating defines model for RateDishReq.Rating.
type RateDishReqRating int

// SearchDishByDateReq Request to look up all dishes served on a date optionally filtered by a location
type SearchDishByDateReq struct {
	// Date Date on which dishes must have been served. Format YYYY-MM-DD
	Date openapi_types.Date `json:"date"`

	// Location Location by which dishes must have been served
	Location *string `json:"location,omitempty"`
}

// SearchDishReq Request to lookup a dishID by the dish name
type SearchDishReq struct {
	// DishName Dish to search for
	DishName string `json:"dishName"`

	// ServedAt Name of the location where this dish is served
	ServedAt string `json:"servedAt"`
}

// SearchDishResp Contains the dishID the requested dish
type SearchDishResp struct {
	// DishID ID of the searched dish if it was found. Omitted otherwise
	DishID *int64 `json:"dishID,omitempty"`

	// DishName Name of the searched ish
	DishName interface{} `json:"dishName"`

	// FoundDish True if the dish was found
	FoundDish bool `json:"foundDish"`
}

// PostDishesDishIDJSONRequestBody defines body for PostDishesDishID for application/json ContentType.
type PostDishesDishIDJSONRequestBody = RateDishReq

// PostMergedDishesJSONRequestBody defines body for PostMergedDishes for application/json ContentType.
type PostMergedDishesJSONRequestBody = CreateMergedDishReq

// PatchMergedDishesMergedDishIDJSONRequestBody defines body for PatchMergedDishesMergedDishID for application/json ContentType.
type PatchMergedDishesMergedDishIDJSONRequestBody = MergedDishUpdateReq

// PostSearchDishJSONRequestBody defines body for PostSearchDish for application/json ContentType.
type PostSearchDishJSONRequestBody = SearchDishReq

// PostSearchDishByDateJSONRequestBody defines body for PostSearchDishByDate for application/json ContentType.
type PostSearchDishByDateJSONRequestBody = SearchDishByDateReq
