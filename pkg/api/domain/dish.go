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

// ratings is a helper slice containing all valid ratings
var ratings []Rating = []Rating{OneStar, TwoStars, ThreeStars, FourStars, FiveStars}

type Dish struct {
	//Name of this dish
	Name string
	//occurences stores the Dates on which this dish was served
	occurences []time.Time
	//ratings stores the ratings for this dish. Index 0 means 1 star and so on
	ratings map[Rating]int
}

// NewDishToday creates a new dish that was first served today
func NewDishToday(name string) Dish {
	return Dish{
		Name:       name,
		occurences: []time.Time{time.Now()},
		ratings:    make(map[Rating]int, 0),
	}
}

func UnmarshalFromDB(name string, occurrences []time.Time, ratings map[Rating]int) *Dish {
	return &Dish{
		Name:       name,
		occurences: occurrences,
		ratings:    ratings,
	}
}

// Rate stores the given rating
func (d *Dish) Rate(r Rating) {
	val, ok := d.ratings[r]
	if !ok {
		val = 1
	} else {
		val = val + 1
	}
	d.ratings[r] = val
}

// AverageRating returns the average rating or an error if no ratings exist yet
func (d *Dish) AverageRating() (float32, error) {
	ratingSum := float64(0)
	totalCount := float64(0)
	for _, r := range ratings {
		ratingSum += float64(d.ratings[r]) * float64(r)
		totalCount += float64(d.ratings[r])
	}
	if totalCount == 0 {
		return 0, fmt.Errorf("no votes yet")
	}
	return float32(ratingSum / totalCount), nil
}

// Ratings returns all ratings for this dish
func (d *Dish) Ratings() map[Rating]int {
	return d.ratings
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

// WasServedToday registers that this dish was served today.
// Calling this multiple times on the same day has no effect
func (d *Dish) WasServedToday() {
	//only add one entry per day
	if onSameDay(d.occurences[len(d.occurences)-1], time.Now()) {
		return
	}

	d.occurences = append(d.occurences, time.Now())
}
