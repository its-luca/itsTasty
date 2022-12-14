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

func (h *HttpServer) PostSearchDishByDate(ctx context.Context, request PostSearchDishByDateRequestObject) (PostSearchDishByDateResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	matchingDishes, err := h.repo.GetDishByDate(dbCtx, request.Body.Date.Time, request.Body.Location)
	if err != nil {
		log.Printf("GetDishByDate for date %v and location %v failed : %v", request.Body.Date.Time, request.Body.Location, err)
		return PostSearchDishByDate500JSONResponse{}, nil
	}

	return PostSearchDishByDate200JSONResponse(matchingDishes), nil
}

func (h *HttpServer) GetUsersMe(ctx context.Context, _ GetUsersMeRequestObject) (GetUsersMeResponseObject, error) {
	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return GetUsersMe500JSONResponse{}, nil
	}

	return GetUsersMe200JSONResponse{Email: userEmail}, nil
}

func NewHttpServer(repo domain.DishRepo) *HttpServer {
	return &HttpServer{repo: repo}
}

func (h *HttpServer) GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) (GetDishesDishIDResponseObject, error) {
	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return GetDishesDishID500JSONResponse{}, nil
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	dbErrGroup, dbErrGroupCtx := errgroup.WithContext(dbCtx)

	//Fill the following vars with background jobs
	var dishRatingOfUser *domain.DishRating
	var dish *domain.Dish
	var dishRatings *domain.DishRatings
	dishNotFound := false
	dishRatingsNotFound := false

	//Get Rating Of User for Dish
	dbErrGroup.Go(func() error {
		var err error
		dishRatingOfUser, err = h.repo.GetRating(dbErrGroupCtx, userEmail, request.DishID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				dishRatingOfUser = nil
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
			if errors.Is(err, domain.ErrNotFound) {
				dishRatingsNotFound = true
				return nil
			}
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
			return GetDishesDishID404Response{}, nil
		}
		//N.B. that the dish was found
		if dishRatingsNotFound {
			log.Printf("Did not find dish ratings allthough dish exists")
			return GetDishesDishID500JSONResponse{}, nil
		}
		return GetDishesDishID500JSONResponse{}, nil
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
		ServedAt:          dish.ServedAt,
		OccurrenceCount:   len(dish.Occurrences()),
		RatingOfUser:      nil, //updated below if data is available
		Ratings:           ratings,
		RecentOccurrences: recentOccurrences,
	}

	avgRating, err := dishRatings.AverageRating()
	if err != nil {
		if !errors.Is(err, domain.ErrNoVotes) {
			log.Printf("dishRatings.AverageRating : %v", err)
			return GetDishesDishID500JSONResponse{}, nil
		}
		//ErrNoVotes is fine, we simply don't add the average rating to the result
	} else {
		response.AvgRating = &avgRating
	}

	if dishRatingOfUser != nil {
		v := GetDishRespRatingOfUser(dishRatingOfUser.Value)
		response.RatingOfUser = &v
	}

	return response, nil

}

func (h *HttpServer) PostDishesDishID(ctx context.Context, request PostDishesDishIDRequestObject) (PostDishesDishIDResponseObject, error) {

	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return PostDishesDishID500JSONResponse{}, nil
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

	if _, err := h.repo.SetOrCreateRating(dbCtx, userEmail, request.DishID, dishRating); err != nil {
		log.Printf("SetRating for dishID %v by user %v : %v", request.DishID, userEmail, err)
		if errors.Is(err, domain.ErrNotFound) {
			return PostDishesDishID404Response{}, nil
		}
		return PostDishesDishID500JSONResponse{}, nil
	}

	return PostDishesDishID200Response{}, nil

}

func (h *HttpServer) GetGetAllDishes(ctx context.Context, _ GetGetAllDishesRequestObject) (GetGetAllDishesResponseObject, error) {

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	dishIDs, err := h.repo.GetAllDishIDs(dbCtx)
	if err != nil {
		log.Printf("GetAllDishes : %v", err)
		return GetGetAllDishes500JSONResponse{}, nil
	}

	return GetGetAllDishes200JSONResponse(dishIDs), nil
}

func (h *HttpServer) PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) (PostSearchDishResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()
	_, dishID, err := h.repo.GetDishByName(dbCtx, request.Body.DishName, request.Body.ServedAt)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return PostSearchDish200JSONResponse{
				DishName:  request.Body.DishName,
				FoundDish: false,
			}, nil
		}
		log.Printf("GetDishByName for %v: %v", request.Body.DishName, err)
		return PostSearchDish500JSONResponse{}, nil
	}

	return PostSearchDish200JSONResponse{
		DishID:    &dishID,
		DishName:  request.Body.DishName,
		FoundDish: true,
	}, nil
}
