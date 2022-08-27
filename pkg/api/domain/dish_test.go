package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestDishToday() Dish {
	return NewDishToday("testDish")
}

func TestDish_AverageRating(t *testing.T) {
	tests := []struct {
		name      string
		setupDish func() Dish
		want      float32
		wantErr   bool
	}{
		{
			name:      "No ratings",
			setupDish: func() Dish { return newTestDishToday() },
			wantErr:   true,
		},
		{
			name: "Test average",
			setupDish: func() Dish {
				d := newTestDishToday()
				d.Rate(OneStar)
				d.Rate(FiveStars)
				d.Rate(FourStars)
				d.Rate(FourStars)
				return d
			},
			want: 3.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupDish()
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
	d.WasServedToday()
	d.WasServedToday()

	assert.Equalf(t, 1, len(d.Occurrences()), "Calling was served today on the same date should not add occurences")
}
