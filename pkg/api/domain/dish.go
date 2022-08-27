package domain

import (
	"fmt"
	"time"
)

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
	Who   string
	Value Rating
	When  time.Time
}

type DishRatings struct {
	Subject Dish
	ratings []DishRating
}

func NewDishRatings(dish Dish, ratings []DishRating) DishRatings {
	return DishRatings{
		Subject: dish,
		ratings: ratings,
	}
}

// AverageRating returns the average rating or an error if no ratings exist yet
func (d *DishRatings) AverageRating() (float32, error) {
	ratingSum := float64(0)
	totalCount := float64(0)
	for _, r := range d.ratings {
		ratingSum += float64(r.Value)
		totalCount += 1
	}
	if totalCount == 0 {
		return 0, fmt.Errorf("no votes yet")
	}
	return float32(ratingSum / totalCount), nil
}

// Ratings returns all ratings for this dish
func (d *DishRatings) Ratings() map[Rating]int {
	res := make(map[Rating]int)

	for _, v := range d.ratings {
		res[v.Value] = res[v.Value] + 1
	}

	return res
}

type Dish struct {
	//Name of this dish
	Name string
	//occurences stores the Dates on which this dish was served
	occurences []time.Time
}

// NewDishToday creates a new dish that was first served today
func NewDishToday(name string) Dish {
	return Dish{
		Name:       name,
		occurences: []time.Time{time.Now()},
	}
}

// Occurrences returns all dates on which this dish was served
func (d *Dish) Occurrences() []time.Time {
	return d.occurences
}

func onSameDay(t1, t2 time.Time) bool {
	if t1.Year() != t2.Year() {
		return false
	}
	if t1.Month() != t2.Month() {
		return false
	}
	if t1.Day() != t2.Day() {
		return false
	}
	return true
}

// MarkAsServedToday registers that this dish was served today.
// Calling this multiple times on the same day has no effect
func (d *Dish) MarkAsServedToday() {
	//only add one entry per day
	if onSameDay(d.occurences[len(d.occurences)-1], time.Now()) {
		return
	}

	d.occurences = append(d.occurences, time.Now())
}
