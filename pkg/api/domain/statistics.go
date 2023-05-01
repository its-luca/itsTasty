package domain

import (
	"errors"
	"sort"
	"time"
)

type UsersOnVacation struct {
	vacations map[DayPrecisionTime]map[string]interface{}
}

func (u UsersOnVacation) UserHasVacation(user string, day DayPrecisionTime) bool {
	v, ok := u.vacations[day]
	if !ok {
		return false
	}

	_, ok = v[user]
	return ok
}

// NewUsersOnVacation expects a map from dates to users who have vacation on that date
func NewUsersOnVacation(vacations map[DayPrecisionTime]map[string]interface{}) UsersOnVacation {
	return UsersOnVacation{vacations: vacations}
}

type RatingStreak struct {
	Begin DayPrecisionTime
	End   DayPrecisionTime
}

// wholeGroupHadVacation is a helper function returning true if *ALL* users in group had vacation on the given day
func wholeGroupHadVacation(group map[string]interface{}, vacations UsersOnVacation, day DayPrecisionTime) bool {
	for k := range group {
		if !vacations.UserHasVacation(k, day) {
			return false
		}
	}
	return true
}

// groupMemberRated is a helper function returning true, if *AT LEAST ONE* group member has rated on the given day
func groupMemberRated(r []DishRating, group map[string]interface{}) bool {
	for _, rating := range r {
		if _, ok := group[rating.Who]; ok {
			return true
		}
	}
	return false
}

func NewRatingStreakFromDB(begin, end DayPrecisionTime) RatingStreak {
	return RatingStreak{
		Begin: begin,
		End:   end,
	}
}

var ErrNoStreak = errors.New("no streak")

// NewRatingStreak calculates the length of the rating streak of group for the given ratings
// Days that are not workdays or on which the whole group had vacation do not break the streak
// Marker errors: ErrNoStreak
func NewRatingStreak(today DayPrecisionTime, ratings []DishRating, vacations UsersOnVacation,
	isHolidayOrWeekend map[DayPrecisionTime]bool, group map[string]interface{}) (RatingStreak, error) {

	if len(group) == 0 {
		return RatingStreak{}, ErrNoStreak
	}

	//sort slice such that ratings[0] is the most recent rating/vote
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].RatingWhen.After(ratings[j].RatingWhen)
	})

	//preprocessing: group ratings by date
	ratingsByDate := make(map[DayPrecisionTime][]DishRating)
	for _, v := range ratings {
		if ratingsByDate[NewDayPrecisionTime(v.RatingWhen)] == nil {
			ratingsByDate[NewDayPrecisionTime(v.RatingWhen)] = []DishRating{}
		}
		ratingsByDate[NewDayPrecisionTime(v.RatingWhen)] = append(ratingsByDate[NewDayPrecisionTime(v.RatingWhen)], v)
	}

	streakStart := today
	for {
		ratingsToday, ok := ratingsByDate[streakStart]
		//no ratings or no one from group rated
		if !ok || !groupMemberRated(ratingsToday, group) {
			//check if everyone had vacation/holiday
			if wholeGroupHadVacation(group, vacations, streakStart) || isHolidayOrWeekend[streakStart] {
				streakStart = streakStart.PrevDay()
				continue
			}
			//nope -> streak ends
			break
		}

		streakStart = streakStart.PrevDay()
	}
	streakStart = streakStart.NextDay()

	if streakStart.After(today.Time) {
		return RatingStreak{}, ErrNoStreak
	}

	return RatingStreak{
		Begin: streakStart,
		End:   today,
	}, nil

}

// LengthInDays of the streak. A streak that starts and ends on the same day has length 1
func (r RatingStreak) LengthInDays() int {
	return int(r.End.Time.Sub(r.Begin.Time)/(24*time.Hour)) + 1
}
