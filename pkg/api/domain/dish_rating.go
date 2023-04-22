package domain

import (
	"errors"
	"fmt"
	"time"
)

var ErrNoVotes = errors.New("no votes yet")

type Rating int

const (
	OneStar Rating = iota + 1
	TwoStars
	ThreeStars
	FourStars
	FiveStars
)

func NewRatingFromInt(r int) (Rating, error) {
	if r < int(OneStar) || r > int(FiveStars) {
		return OneStar, fmt.Errorf("invalid rating, must be in range [%v;%v]", OneStar, FiveStars)
	}
	return Rating(r), nil
}

type DishRating struct {
	Who        string
	Value      Rating
	RatingWhen time.Time
}

func NewDishRating(who string, rating Rating, when time.Time) DishRating {
	return DishRating{
		Who:        who,
		Value:      rating,
		RatingWhen: when,
	}
}

func NewDishRatingFromDB(who string, rating int, when time.Time) (DishRating, error) {
	r, err := NewRatingFromInt(rating)
	if err != nil {
		return DishRating{}, err
	}
	return DishRating{
		Who:        who,
		Value:      r,
		RatingWhen: when,
	}, nil
}

// AverageRating returns the average rating or an error if no ratings exist yet
func AverageRating(ratings []DishRating) (float32, error) {
	ratingSum := float64(0)
	totalCount := float64(0)
	for _, r := range ratings {
		ratingSum += float64(r.Value)
		totalCount += 1
	}
	if totalCount == 0 {
		return 0, ErrNoVotes
	}
	return float32(ratingSum / totalCount), nil
}

// Ratings returns all ratings for this dish
func Ratings(ratings []DishRating) map[Rating]int {
	res := make(map[Rating]int)

	for _, v := range ratings {
		res[v.Value] = res[v.Value] + 1
	}

	return res
}
