package botAPI

import (
	"context"
	"errors"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"golang.org/x/sync/errgroup"
	"itsTasty/pkg/api/domain"
	"log"
	"time"
)

//go:generate oapi-codegen --config ./server.cfg.yml ../../botAPI.yml
//go:generate oapi-codegen --config ./types.cfg.yml ../../botAPI.yml

const defaultDBTimeout = 5 * time.Second

type Service struct {
	repo domain.DishRepo
}

func NewService(repo domain.DishRepo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) PostCreateOrUpdateDish(ctx context.Context, request PostCreateOrUpdateDishRequestObject) (PostCreateOrUpdateDishResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	_, createdDish, createdLocation, dishID, err := s.repo.GetOrCreateDish(dbCtx, request.Body.DishName, request.Body.ServedAt)
	if err != nil {
		log.Printf("GetOrCreateDish for dishName %v : %v", request.Body.DishName, err)
		return PostCreateOrUpdateDish500JSONResponse{}, nil
	}
	dbCancel()

	dbCtx, dbCancel = context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	err = s.repo.UpdateMostRecentServing(dbCtx, dishID, func(currenMostRecent *time.Time) (*time.Time, error) {
		if currenMostRecent == nil {
			newMostRecentServing := domain.NowWithDayPrecision()
			return &newMostRecentServing, nil
		}
		if !domain.OnSameDay(*currenMostRecent, time.Now()) {
			newMostRecentServing := domain.NowWithDayPrecision()
			return &newMostRecentServing, nil
		}
		return nil, nil
	})
	if err != nil {
		log.Printf("UPdateDish for dishID %v : %v", dishID, err)
		return PostCreateOrUpdateDish500JSONResponse{}, nil
	}

	//
	// Assemble Response
	//

	return PostCreateOrUpdateDish200JSONResponse{
		CreatedNewDish:     createdDish,
		CreatedNewLocation: createdLocation,
		DishID:             dishID,
	}, nil

}

func (s *Service) GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) (GetDishesDishIDResponseObject, error) {

	//
	// Query data
	//

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	dbErrGroup, dbErrGroupCtx := errgroup.WithContext(dbCtx)

	var dish *domain.Dish
	var dishRatings *domain.DishRatings
	dishNotFound := false

	//Get all Ratings for Dish
	dbErrGroup.Go(func() error {
		var err error
		dishRatings, err = s.repo.GetAllRatingsForDish(dbErrGroupCtx, request.DishID)
		if err != nil {
			return fmt.Errorf("GetAllRatingsForDish : %v", err)
		}
		return nil
	})

	//Get dish
	dbErrGroup.Go(func() error {
		var err error
		dish, err = s.repo.GetDishByID(dbErrGroupCtx, request.DishID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				dishNotFound = true
			}
			return fmt.Errorf("GetDishByID : %v", err)
		}
		return nil
	})

	//Wait for all jobs and check if there was an error
	err := dbErrGroup.Wait()
	if err != nil {
		log.Printf("Job in errgroup failed : %v", err)
		if dishNotFound {
			return GetDishesDishID404Response{}, nil
		}
		return GetDishesDishID500JSONResponse{}, nil
	}

	//
	//Assemble response
	//

	const maxOccurrencesInAnswer = 10
	occurrences := dish.Occurrences()
	occurrenceCountForResponse := maxOccurrencesInAnswer
	if maxOccurrencesInAnswer > len(occurrences) {
		occurrenceCountForResponse = len(occurrences)
	}
	recentOccurrences := make([]types.Date, 0, occurrenceCountForResponse)
	for i := range occurrences {
		recentOccurrences = append(recentOccurrences, types.Date{Time: occurrences[i]})
	}

	ratings := make(map[string]int)
	for k, v := range dishRatings.Ratings() {
		ratings[fmt.Sprintf("%v", k)] = v
	}

	response := GetDishesDishID200JSONResponse{
		AvgRating:         nil, //updated below if data is available
		Name:              dish.Name,
		OccurrenceCount:   len(dish.Occurrences()),
		Ratings:           ratings,
		RecentOccurrences: recentOccurrences,
		ServedAt:          dish.ServedAt,
	}

	if avgRating, err := dishRatings.AverageRating(); err == nil {
		response.AvgRating = &avgRating
	}

	return response, nil
}
