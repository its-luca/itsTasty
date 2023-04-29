package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestNewRatingStreak(t *testing.T) {

	today := NewDayPrecisionTime(time.Now())
	type args struct {
		today              DayPrecisionTime
		ratings            []DishRating
		vacations          UsersOnVacation
		isHolidayOrWeekend map[DayPrecisionTime]bool
		group              map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want RatingStreak
	}{
		{
			name: "Empty Group",
			args: args{
				today: today,
				ratings: []DishRating{
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.Time,
					},
				},
				vacations: UsersOnVacation{
					vacations: map[DayPrecisionTime]map[string]interface{}{
						{
							Time: time.Time{},
						}: map[string]interface{}{
							"": nil,
						},
					},
				},
				isHolidayOrWeekend: map[DayPrecisionTime]bool{
					{
						Time: time.Time{},
					}: false,
				},
				group: map[string]interface{}{},
			},
			want: RatingStreak{
				Begin: today,
				End:   today,
			},
		},
		{
			name: "3 simple, consecutive Ratings",
			args: args{
				today: today,
				ratings: []DishRating{
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.Time,
					},
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.PrevDay().Time,
					},
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.PrevDay().PrevDay().Time,
					},
				},
				vacations: UsersOnVacation{
					vacations: map[DayPrecisionTime]map[string]interface{}{
						{
							Time: time.Time{},
						}: map[string]interface{}{
							"": nil,
						},
					},
				},
				isHolidayOrWeekend: map[DayPrecisionTime]bool{
					{
						Time: time.Time{},
					}: false,
				},
				group: map[string]interface{}{
					"user1": nil,
				},
			},
			want: RatingStreak{
				Begin: today.PrevDay().PrevDay().PrevDay(),
				End:   today,
			},
		},
		{
			name: "Two user group, 3 ratings, interrupted by one vacation day and one other non work day",
			args: args{
				today: today,
				ratings: []DishRating{
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.Time,
					},
					//N.B: one day gap, user had vacation on that day
					{
						Who:        "user1",
						Value:      1,
						RatingWhen: today.PrevDay().PrevDay().Time,
					},
					//N.B one day gap, was a non working day
					{
						Who:        "user2",
						Value:      1,
						RatingWhen: today.PrevDay().PrevDay().PrevDay().PrevDay().Time,
					},
				},
				vacations: UsersOnVacation{
					vacations: map[DayPrecisionTime]map[string]interface{}{
						{
							Time: today.PrevDay().Time,
						}: {
							"user1": nil,
							"user2": nil,
						},
					},
				},
				isHolidayOrWeekend: map[DayPrecisionTime]bool{
					{
						Time: today.PrevDay().PrevDay().PrevDay().Time,
					}: true,
				},
				group: map[string]interface{}{
					"user1": nil,
					"user2": nil,
				},
			},
			want: RatingStreak{
				Begin: today.PrevDay().PrevDay().PrevDay().PrevDay().PrevDay(),
				End:   today,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRatingStreak(tt.args.today, tt.args.ratings, tt.args.vacations, tt.args.isHolidayOrWeekend, tt.args.group); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRatingStreak() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRatingStreak_LengthInDays(t *testing.T) {
	today := NewDayPrecisionTime(time.Now())
	type fields struct {
		Begin DayPrecisionTime
		End   DayPrecisionTime
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "Empty Rating Streak",
			fields: fields{
				Begin: today,
				End:   today,
			},
			want: 0,
		},
		{
			name: "1 day Rating Streak",
			fields: fields{
				Begin: today.PrevDay(),
				End:   today,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RatingStreak{
				Begin: tt.fields.Begin,
				End:   tt.fields.End,
			}
			if got := r.LengthInDays(); got != tt.want {
				t.Errorf("RatingStreak.LengthInDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
