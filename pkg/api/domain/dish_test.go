package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func newTestDishToday() Dish {
	return NewDishToday("testDish")
}

func TestDish_AverageRating(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() DishRatings
		want    float32
		wantErr bool
	}{
		{
			name:    "No ratings",
			setup:   func() DishRatings { return NewDishRatings(newTestDishToday(), make([]DishRating, 0)) },
			wantErr: true,
		},
		{
			name: "Test average",
			setup: func() DishRatings {
				d := newTestDishToday()
				ratings := []DishRating{
					{
						Who:   "userA",
						Value: OneStar,
						When:  time.Time{},
					},
					{
						Who:   "userB",
						Value: FiveStars,
						When:  time.Time{},
					},
					{
						Who:   "userC",
						Value: FourStars,
						When:  time.Time{},
					},
					{
						Who:   "userD",
						Value: FourStars,
						When:  time.Time{},
					},
				}
				return NewDishRatings(d, ratings)
			},
			want: 3.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			got, err := d.AverageRating()
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

	d := NewDishToday("testDish")
	d.MarkAsServedToday()
	d.MarkAsServedToday()

	assert.Equalf(t, 1, len(d.Occurrences()), "Calling was served today on the same date should not add occurences")
}
