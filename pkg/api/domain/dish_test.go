package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func newTestDishToday() Dish {
	return *NewDishToday("testDish", "testLocation")
}

func TestDish_AverageRating(t *testing.T) {
	tests := []struct {
		name    string
		input   []DishRating
		want    float32
		wantErr bool
	}{
		{
			name:    "No ratings",
			input:   make([]DishRating, 0),
			wantErr: true,
		},
		{
			name: "Test average",
			input: []DishRating{
				{
					Who:        "userA",
					Value:      OneStar,
					RatingWhen: time.Time{},
				},
				{
					Who:        "userB",
					Value:      FiveStars,
					RatingWhen: time.Time{},
				},
				{
					Who:        "userC",
					Value:      FourStars,
					RatingWhen: time.Time{},
				},
				{
					Who:        "userD",
					Value:      FourStars,
					RatingWhen: time.Time{},
				},
			},
			want: 3.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AverageRating(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dish.AverageRating() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Dish.AverageRating() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDish_WasServedToday(t *testing.T) {

	d := NewDishToday("testDish", "testLocation")

	now := time.Now()
	d.UpdateOccurrenceIfNewDay(now)
	d.UpdateOccurrenceIfNewDay(now)

	assert.Equalf(t, 1, len(d.Occurrences()), "Calling UpdateOccurrenceIfNewDay on the same date should not add occurences")

	d.UpdateOccurrenceIfNewDay(now.Add(-24 * time.Hour))
	assert.Equalf(t, 1, len(d.Occurrences()), "Calling UpdateOccurrenceIfNewDay with a date before the most recent serving should not add occurences")

	d.UpdateOccurrenceIfNewDay(now.Add(24 * time.Hour))
	assert.Equalf(t, 2, len(d.Occurrences()), "Calling UpdateOccurrenceIfNewDay with a date that is at least one day ine the future SHOULD ADD a new occurence")

}
