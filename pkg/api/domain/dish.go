package domain

import (
	"sort"
	"time"
)

type Dish struct {
	//Name of this dish
	Name string
	//ServedAt is the location where the dish is served
	ServedAt string
	//occurences stores the Dates on which this dish was served
	occurences []time.Time
}

// NewDishToday creates a new dish that was first served today
func NewDishToday(name string, servedAt string) *Dish {
	return &Dish{
		Name:       name,
		ServedAt:   servedAt,
		occurences: []time.Time{TruncateToDayPrecision(time.Now())},
	}
}

func NewDishFromDB(name, servedAt string, occurrences []time.Time) *Dish {
	sort.Slice(occurrences, func(i, j int) bool {
		return occurrences[i].Before(occurrences[j])
	})
	return &Dish{
		Name:       name,
		ServedAt:   servedAt,
		occurences: occurrences,
	}
}

// Occurrences returns all dates on which this dish was served in ascending order
// (i.e. index 0 holds the oldest serving)
func (d *Dish) Occurrences() []time.Time {
	return d.occurences
}

// CreateNewRatingInsteadOfUpdating returns true if the user is allowed to create a new rating
// Otherwise he must update his most recent rating. mostRecentRating may be nil
func (d *Dish) CreateNewRatingInsteadOfUpdating(mostRecentRating *DishRating, newRating DishRating) bool {
	//if there is no rating, create a new one
	if mostRecentRating == nil {
		return true
	}

	//otherwise, only create a new rating if there has been a new occurrence since the most recent rating

	//newRating is on same day or earlier than mostRecentRating -> update as their cannot be a new occurrence
	if OnSameDay(mostRecentRating.RatingWhen, newRating.RatingWhen) || newRating.RatingWhen.Before(mostRecentRating.RatingWhen) {
		return false
	}

	//if we are here, newRating is at least on the next day after mostRecentRating. Check if there has been a new occurrence since
	mostRecentOccurrence := d.Occurrences()[len(d.Occurrences())-1]
	return mostRecentOccurrence.After(mostRecentRating.RatingWhen) && !OnSameDay(mostRecentOccurrence, mostRecentRating.RatingWhen)

}

/*TODO: UpdateOccurenceIfNewDay is currently not connected to a backend updated function
updating the whole Dish object at once is awkward. Currently there is a dedicated function
to update serving/occurrences
*/

// UpdateOccurrenceIfNewDay adds a new serving at time t if it is at least one day has passed since the last serving
// (i.e. when called multiple times for the same day, at most one serving is added). Dates that are prior to the
// current most recent serving are ignored
func (d *Dish) UpdateOccurrenceIfNewDay(t time.Time) {

	//we enforce this in the constructors
	if len(d.occurences) == 0 {
		panic("dish objects must always have at least one serving")
	}

	tDayPrec := TruncateToDayPrecision(t)
	currDayPrec := TruncateToDayPrecision(d.occurences[len(d.occurences)-1])

	//only allow future servings
	if tDayPrec.Before(currDayPrec) {
		return
	}

	if OnSameDay(d.occurences[len(d.occurences)-1], t) {
		return
	}

	d.occurences = append(d.occurences, TruncateToDayPrecision(t))
}
