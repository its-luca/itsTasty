package ports

import (
	"context"
	"errors"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
	"itsTasty/pkg/api/domain"
	"time"
)

type DishWithRatings struct {
	Dish    *domain.Dish
	DishID  int64
	Ratings []domain.DishRating
}

type BasicDishReply struct {
	AvgRating         *float32
	Name              string
	OccurrenceCount   int
	Ratings           map[string]int
	RecentOccurrences []types.Date
	ServedAt          string
}

func FetchDishResources(ctx context.Context, repo domain.DishRepo, dishName, servedAt string) (*DishWithRatings, error) {
	dish, dishID, err := repo.GetDishByName(ctx, dishName, servedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Dish : %w", err)
	}

	ratings, err := repo.GetAllRatingsForDish(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Ratings : %w", err)
	}

	return &DishWithRatings{
		Dish:    dish,
		DishID:  dishID,
		Ratings: ratings,
	}, nil

}

func FetchDishResourcesByID(ctx context.Context, repo domain.DishRepo, dishID int64) (*DishWithRatings, error) {

	p := pool.New().WithContext(ctx).WithCancelOnError()

	result := &DishWithRatings{
		Dish:    nil, //filled by goroutine
		DishID:  dishID,
		Ratings: nil, //filled by goroutine
	}

	p.Go(func(ctx context.Context) error {
		dish, err := repo.GetDishByID(ctx, dishID)
		if err != nil {
			return fmt.Errorf("failed to fetch Dish : %w", err)
		}
		result.Dish = dish
		return nil
	})

	p.Go(func(ctx context.Context) error {
		ratings, err := repo.GetAllRatingsForDish(ctx, dishID)
		if err != nil {
			return fmt.Errorf("failed to fetch Ratings : %w", err)
		}
		result.Ratings = ratings
		return nil
	})

	if err := p.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func FetchBasicDishData(ctx context.Context, repo domain.DishRepo, dishID int64) (*BasicDishReply, error) {

	isPartOfMergedDish, mergedDishID, err := repo.IsDishPartOfMergedDisByID(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("IsDishPartOfMergedDisByID failed : %v", err)
	}

	dataForResponse := make([]*DishWithRatings, 0)
	var mergedDish *domain.MergedDish

	if isPartOfMergedDish {
		mergedDish, err = repo.GetMergedDishByID(ctx, mergedDishID)
		if err != nil {

			return nil, fmt.Errorf("failed to check if dish %v is part of a merged dish : %v",
				dishID, err)
		}
		mapper := iter.Mapper[string, *DishWithRatings]{MaxGoroutines: 3}
		dishData, err := mapper.MapErr(mergedDish.GetCondensedDishNames(), func(dishName *string) (*DishWithRatings, error) {
			return FetchDishResources(ctx, repo, *dishName, mergedDish.ServedAt)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch at lesat one dish contained in merged dish (name=%v, id=%v) :  %v", mergedDish.Name, mergedDishID, err)
		}
		dataForResponse = append(dataForResponse, dishData...)
	} else {
		dishData, err := FetchDishResourcesByID(ctx, repo, dishID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch dish %v: %v", dishID, err)
		}
		dataForResponse = append(dataForResponse, dishData)
	}

	//
	//Assemble response
	//

	//occurrence data

	const maxOccurrencesInAnswer = 10

	occurrences := make([]time.Time, 0)
	if mergedDish == nil {
		occurrences = dataForResponse[0].Dish.Occurrences()
	} else {
		rawOccurrences := make([]time.Time, 0)
		for _, v := range dataForResponse {
			rawOccurrences = append(rawOccurrences, v.Dish.Occurrences()...)
		}
		occurrences = mergedDish.GetUniqueOccurrences(rawOccurrences)
	}
	occurrenceCountForResponse := maxOccurrencesInAnswer
	if maxOccurrencesInAnswer > len(occurrences) {
		occurrenceCountForResponse = len(occurrences)
	}
	recentOccurrences := make([]types.Date, 0, occurrenceCountForResponse)
	for i := range occurrences {
		recentOccurrences = append(recentOccurrences, types.Date{Time: occurrences[i]})
	}

	//rating data

	allDishRatings := make([]domain.DishRating, 0)
	for _, v := range dataForResponse {
		allDishRatings = append(allDishRatings, v.Ratings...)
	}
	var avgRating *float32
	if v, err := domain.AverageRating(allDishRatings); err != nil {
		if !errors.Is(err, domain.ErrNoVotes) {
			return nil, fmt.Errorf("failed to calculate average rating : %v", err)
		}
	} else {
		avgRating = &v
	}
	ratings := make(map[string]int)
	for k, v := range domain.Ratings(allDishRatings) {
		ratings[fmt.Sprintf("%v", k)] = v
	}

	var name string
	if mergedDish == nil {
		name = dataForResponse[0].Dish.Name
	} else {
		name = mergedDish.Name
	}

	var servedAt string
	if mergedDish == nil {
		servedAt = dataForResponse[0].Dish.ServedAt
	} else {
		servedAt = mergedDish.ServedAt
	}

	return &BasicDishReply{
		AvgRating:         avgRating, //updated below if data is available
		Name:              name,
		OccurrenceCount:   len(occurrences),
		Ratings:           ratings,
		RecentOccurrences: recentOccurrences,
		ServedAt:          servedAt,
	}, nil
}
