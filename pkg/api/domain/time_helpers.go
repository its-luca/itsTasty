package domain

import (
	"time"
)

type DayPrecisionTime struct {
	time.Time
}

func NewDayPrecisionTime(t time.Time) DayPrecisionTime {
	t = TruncateToDayPrecision(t)
	return DayPrecisionTime{t}
}

func (d DayPrecisionTime) PrevDay() DayPrecisionTime {
	return NewDayPrecisionTime(d.Add(-24 * (time.Hour)))
}

func (d DayPrecisionTime) NextDy() DayPrecisionTime {
	return NewDayPrecisionTime(d.Add(24 * time.Hour))
}

// OnSameDay returns true if t1 and t2 are on the same day
func OnSameDay(t1, t2 time.Time) bool {
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

// TruncateToDayPrecision returns t with hour, min, sec and nsec set to zero
func TruncateToDayPrecision(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
