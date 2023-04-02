package botAPI

import (
	"context"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports"
	"log"
	"time"
)

//go:generate oapi-codegen --config ./server.cfg.yml ../../botAPI.yml
//go:generate oapi-codegen --config ./types.cfg.yml ../../botAPI.yml

const defaultDBTimeout = 5 * time.Second

type Service struct {
	repo       domain.DishRepo
	timeSource TimeSource
}

// TimeSource allows mock time when testing
type TimeSource interface {
	//Now returns the current local time.
	Now() time.Time
}

// defaultTimeSource simply wraps time.Now()
type defaultTimeSource struct {
}

func (d defaultTimeSource) Now() time.Time {
	return time.Now()
}

func NewService(repo domain.DishRepo) *Service {
	return &Service{
		repo:       repo,
		timeSource: defaultTimeSource{},
	}
}

type ServiceFactory func(repo domain.DishRepo) *Service

func NewServiceCustomTime(repo domain.DishRepo, timeSource TimeSource) *Service {
	return &Service{
		repo:       repo,
		timeSource: timeSource,
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
			newMostRecentServing := domain.TruncateToDayPrecision(s.timeSource.Now())
			return &newMostRecentServing, nil
		}
		if !domain.OnSameDay(*currenMostRecent, s.timeSource.Now()) {
			newMostRecentServing := domain.TruncateToDayPrecision(s.timeSource.Now())
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

	basicDishData, err := ports.FetchBasicDishData(dbCtx, s.repo, request.DishID)
	if err != nil {
		log.Printf("FetchBasicDishData for dishID %v failed : %v", request.DishID, err)
		return GetDishesDishID500JSONResponse{}, nil
	}

	response := GetDishesDishID200JSONResponse{
		AvgRating:         basicDishData.AvgRating, //updated below if data is available
		Name:              basicDishData.Name,
		ServedAt:          basicDishData.ServedAt,
		OccurrenceCount:   basicDishData.OccurrenceCount,
		Ratings:           basicDishData.Ratings,
		RecentOccurrences: basicDishData.RecentOccurrences,
	}

	return response, nil
}
