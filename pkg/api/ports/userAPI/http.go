package userAPI

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

//go:generate oapi-codegen --config ./server.cfg.yml ../../userAPI.yml
//go:generate oapi-codegen --config ./types.cfg.yml ../../userAPI.yml

type contextKey string

const userEmailContextKey = contextKey("userEmail")

func ContextWithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailContextKey, email)
}

func GetUserEmailFromCTX(ctx context.Context) (string, error) {
	userEmail, ok := ctx.Value(userEmailContextKey).(string)
	if !ok {
		return "", fmt.Errorf("failed to cast to string type")
	}
	return userEmail, nil
}

const defaultDBTimeout = 5 * time.Second

type HttpServer struct {
	repo domain.DishRepo
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (h HttpServer) GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) interface{} {
	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return PostDishesDishID500JSONResponse{}
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	dbErrGroup, dbErrGroupCtx := errgroup.WithContext(dbCtx)

	//Fill the following vars with background jobs
	var dishRatignOfUser *domain.DishRating
	var dish *domain.Dish
	var dishRatings *domain.DishRatings
	dishNotFound := false

	//Get Rating Of User for Dish
	dbErrGroup.Go(func() error {
		var err error
		dishRatignOfUser, _, err = h.repo.GetRating(dbErrGroupCtx, userEmail, request.DishID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				dishRatignOfUser = nil
				return nil
			}
			return fmt.Errorf("GetRating : %v", err)
		}
		return nil
	})

	//Get all Ratings for Dish
	dbErrGroup.Go(func() error {
		var err error
		dishRatings, err = h.repo.GetAllRatingsForDish(dbErrGroupCtx, request.DishID)
		if err != nil {
			return fmt.Errorf("GetAllRatingsForDish : %v", err)
		}
		return nil
	})

	//Get dish
	dbErrGroup.Go(func() error {
		var err error
		dish, err = h.repo.GetDishByID(dbErrGroupCtx, request.DishID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				dishNotFound = true
			}
			return fmt.Errorf("GetDishByID : %v", err)
		}
		return nil
	})

	//Wait for all jobs and check if there was an error
	err = dbErrGroup.Wait()
	if err != nil {
		log.Printf("Job in errgroup failed : %v", err)
		if dishNotFound {
			return GetDishesDishID404Response{}
		}
		return GetDishesDishID500JSONResponse{}
	}

	ratings := make(map[string]int)
	for k, v := range dishRatings.Ratings() {
		ratings[fmt.Sprintf("%v", k)] = v
	}

	const maxOccurencesInAnswer = 10
	occurences := dish.Occurrences()
	occurenceCountForResponse := maxOccurencesInAnswer
	if maxOccurencesInAnswer > len(occurences) {
		occurenceCountForResponse = len(occurences)
	}
	recentOccurrences := make([]types.Date, 0, occurenceCountForResponse)
	for i := range occurences {
		recentOccurrences = append(recentOccurrences, types.Date{Time: occurences[i]})
	}

	response := GetDishesDishID200JSONResponse{
		AvgRating:         nil, //updated below if data is available
		Name:              dish.Name,
		OccurrenceCount:   len(dish.Occurrences()),
		RatingOfUser:      nil, //updated below if data is available
		Ratings:           ratings,
		RecentOccurrences: recentOccurrences,
	}

	if avgRating, err := dishRatings.AverageRating(); err != nil {
		response.AvgRating = &avgRating
	}

	if dishRatignOfUser != nil {
		v := GetDishRespRatingOfUser(dishRatignOfUser.Value)
		response.RatingOfUser = &v
	}

	return response

}

func (h HttpServer) PostDishesDishID(ctx context.Context, request PostDishesDishIDRequestObject) interface{} {

	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return PostDishesDishID500JSONResponse{}
	}

	rating, err := domain.NewRatingFromInt(int(request.Body.Rating))
	if err != nil {
		log.Printf("User %v gave invalid rating : %v", userEmail, err)
	}

	dishRating := domain.DishRating{
		Who:   userEmail,
		Value: rating,
		When:  time.Now(),
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	if err := h.repo.SetRating(dbCtx, userEmail, request.DishID, dishRating); err != nil {
		log.Printf("SetRating for dishID %v by user %v : %v", request.DishID, userEmail, err)
		if errors.Is(err, domain.ErrNotFound) {
			return PostDishesDishID404Response{}
		}
		return PostDishesDishID500JSONResponse{}
	}

	return PostDishesDishID200Response{}

}

func (h HttpServer) GetGetAllDishes(ctx context.Context, _ GetGetAllDishesRequestObject) interface{} {

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	dishIDs, err := h.repo.GetAllDishIDs(dbCtx)
	if err != nil {
		log.Printf("GetAllDishes : %v", err)
		return GetGetAllDishes500JSONResponse{}
	}

	return GetGetAllDishes200JSONResponse(dishIDs)
}

func (h HttpServer) PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) interface{} {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	_, dishID, err := h.repo.GetDishByName(dbCtx, request.Body.DishName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return PostSearchDish200JSONResponse{
				DishName:  request.Body.DishName,
				FoundDish: false,
			}
		}
		log.Printf("GetDishByName for %v: %v", request.Body.DishName, err)
		return PostSearchDish500JSONResponse{}
	}

	return PostSearchDish200JSONResponse{
		DishID:    &dishID,
		DishName:  request.Body.DishName,
		FoundDish: true,
	}
}
