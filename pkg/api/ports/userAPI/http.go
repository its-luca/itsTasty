package userAPI

import (
	"context"
	"errors"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/sourcegraph/conc/iter"
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

func (h *HttpServer) GetMergedDishesMergedDishID(ctx context.Context, r GetMergedDishesMergedDishIDRequestObject) (GetMergedDishesMergedDishIDResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	//get merged dish

	mergedDish, err := h.repo.GetMergedDishByID(dbCtx, r.MergedDishID)
	if err != nil {
		log.Printf("failed get merged dish %v : %v", r.MergedDishID, err)
		if errors.Is(err, domain.ErrNotFound) {
			return GetMergedDishesMergedDishID404Response{}, nil
		}
		return GetMergedDishesMergedDishID500Response{}, nil
	}

	//get id for dishes contained in merged dish. This is a convenience feature of the API

	mapper := iter.Mapper[string, int64]{
		MaxGoroutines: 3,
	}
	dishNames := mergedDish.GetCondensedDishNames()
	dishIDs, err := mapper.MapErr(dishNames, func(dishName *string) (int64, error) {
		_, dishID, err := h.repo.GetDishByName(dbCtx, *dishName, mergedDish.ServedAt)
		if err != nil {
			return 0, fmt.Errorf("failed to get dish (name: %v, loation: %v) : %w", *dishName, mergedDish.ServedAt, err)
		}
		return dishID, nil
	})
	if err != nil {
		log.Printf("failed get at least one dish for dishes %v served at %v : %v", dishNames, mergedDish.ServedAt, err)
		return GetMergedDishesMergedDishID500Response{}, nil
	}

	//assemble response

	resp := GetMergedDishesMergedDishID200JSONResponse{
		ContainedDishIDs:   dishIDs,
		ContainedDishNames: dishNames,
		Name:               mergedDish.Name,
		ServedAt:           mergedDish.ServedAt,
	}

	return resp, nil
}

func (h *HttpServer) PostMergedDishes(ctx context.Context, request PostMergedDishesRequestObject) (PostMergedDishesResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	if len(request.Body.MergedDishes) < 2 {
		s := "You must provide at least two dish IDs"
		return PostMergedDishes400JSONResponse{What: &s}, nil
	}

	//TODO: fetch in parallel. take a look at https://github.com/sourcegraph/conc/
	//
	// Fetch Resources
	//

	dishesForMerge := make([]*domain.Dish, 0, len(request.Body.MergedDishes))
	for _, v := range request.Body.MergedDishes {
		d, err := h.repo.GetDishByID(dbCtx, v)
		if err != nil {
			log.Printf("GetDishByID for id %v failed : %v", v, err)
			if errors.Is(err, domain.ErrNotFound) {

				s := fmt.Sprintf("dishID %v does not exist", v)
				return PostMergedDishes400JSONResponse{What: &s}, nil
			}
			return PostMergedDishes500Response{}, nil
		}

		dishesForMerge = append(dishesForMerge, d)
	}

	//
	// Check domain Logic
	//

	mergedDish, err := domain.NewMergedDish(request.Body.Name, dishesForMerge[0], dishesForMerge[1], dishesForMerge[2:])
	if err != nil {
		log.Printf("domain.NewMergedDish failed for request %v failed : %v", request.Body, err)
		if errors.Is(err, domain.ErrNotOnSameLocation) {
			s := "All dishes must be served at the same location"
			return PostMergedDishes400JSONResponse{What: &s}, nil
		}
		return PostMergedDishes500Response{}, nil
	}

	//
	// Success! Persist to backend
	//
	mergedDishID, err := h.repo.CreateMergedDish(dbCtx, mergedDish)
	if err != nil {
		log.Printf("repo.CreateMergedDish failed for  %v failed : %v", mergedDish, err)
		if errors.Is(err, domain.ErrDishAlreadyMerged) {
			s := "One of the dishes is already part of a merged dish"
			return PostMergedDishes400JSONResponse{What: &s}, nil
		}
		return PostMergedDishes500Response{}, nil
	}

	return PostMergedDishes200JSONResponse{
		MergedDishID: mergedDishID,
	}, nil

}

func (h *HttpServer) DeleteMergedDishesMergedDishID(ctx context.Context, r DeleteMergedDishesMergedDishIDRequestObject) (DeleteMergedDishesMergedDishIDResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	err := h.repo.DeleteMergedDishByID(dbCtx, r.MergedDishID)
	if err != nil {
		log.Printf("Failed to delte merged dish %v : %v", r.MergedDishID, err)

		if errors.Is(err, domain.ErrNotFound) {
			return DeleteMergedDishesMergedDishID404Response{}, nil
		}
		return DeleteMergedDishesMergedDishID500Response{}, nil
	}
	return DeleteMergedDishesMergedDishID200Response{}, nil
}

func (h *HttpServer) PatchMergedDishesMergedDishID(ctx context.Context, request PatchMergedDishesMergedDishIDRequestObject) (PatchMergedDishesMergedDishIDResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	//
	// Fetch dishes for adding/removing from merged dish
	//

	mapper := iter.Mapper[int64, *domain.Dish]{
		//TODO: what is a reasonable value?
		MaxGoroutines: 3,
	}
	addDishes := make([]*domain.Dish, 0)
	if addDishIDs := request.Body.AddDishIDs; addDishIDs != nil {
		//fetch dishes for adding
		var err error
		addDishes, err = mapper.MapErr(*addDishIDs, func(i *int64) (*domain.Dish, error) {
			d, err := h.repo.GetDishByID(dbCtx, *i)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch dish id %v : %w", *i, err)
			}
			return d, nil
		})
		if err != nil {
			log.Printf("failed to fetch dishes %v for adding to merged dish %v : %v", addDishIDs, request.MergedDishID, err)
			if errors.Is(err, domain.ErrNotFound) {
				s := "At least one of the provided dishes that should be added could not be found"
				return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
			}
			return PatchMergedDishesMergedDishID500Response{}, nil
		}
	}

	removeDishes := make([]*domain.Dish, 0)
	if removeDishIDs := request.Body.RemoveDishIDs; removeDishIDs != nil {
		//fetch dishes for removal
		var err error
		removeDishes, err = mapper.MapErr(*removeDishIDs, func(i *int64) (*domain.Dish, error) {
			d, err := h.repo.GetDishByID(dbCtx, *i)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch dish id %v : %w", *i, err)
			}
			return d, nil
		})
		if err != nil {
			log.Printf("failed to fetch dishes %v for removal from merged dish %v : %v", removeDishes, request.MergedDishID, err)
			if errors.Is(err, domain.ErrNotFound) {
				s := "At least one of the  dishes that should be removed could not be found"
				return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
			}
			return PatchMergedDishesMergedDishID500Response{}, nil
		}
	}

	//
	// Update merged dish
	//

	//add dishes
	err := h.repo.UpdateMergedDishByID(dbCtx, request.MergedDishID, func(current *domain.MergedDish) (*domain.MergedDish, error) {
		for _, v := range addDishes {
			if err := current.AddDish(v); err != nil {
				return nil, fmt.Errorf("cannot add dish %v : %w", *v, err)
			}
		}

		for _, v := range removeDishes {
			if err := current.RemoveDish(v); err != nil {
				return nil, fmt.Errorf("cannot remove dish %v  : %w", *v, err)
			}
		}

		return current, nil
	})
	if err != nil {

		log.Printf("UpdateMergedDishByID failed to add dishes %v to merged dish id %v : %v", request.MergedDishID, addDishes, err)
		//common errors

		if errors.Is(err, domain.ErrNotFound) {
			return PatchMergedDishesMergedDishID404Response{}, nil
		}

		//add errors

		if errors.Is(err, domain.ErrNotOnSameLocation) {
			s := "at least one of the dishes is not served on the same location as the merged dish"
			return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
		}

		if errors.Is(err, domain.ErrDishAlreadyMerged) {
			s := "at least one of the dishes is already part of a merged dish"
			return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
		}

		//remove errors

		if errors.Is(err, domain.ErrDishNotPartOfMergedDish) {
			s := "at least one of the dishes that should be removed is not part of the merged dish"
			return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
		}

		if errors.Is(err, domain.ErrMergedDishNeedsAtLeastTwoDishes) {
			s := "after removal, the merged dish would have less than two dishes left"
			return PatchMergedDishesMergedDishID400JSONResponse{What: &s}, nil
		}

		//unexpected errors
		return PatchMergedDishesMergedDishID500Response{}, nil
	}

	return PatchMergedDishesMergedDishID200Response{}, nil
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
		var ratings []domain.DishRating
		ratings, err = h.repo.GetRatings(dbErrGroupCtx, userEmail, request.DishID, true)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				dishRatingOfUser = nil
				return nil
			}
			if len(ratings) != 1 {
				return fmt.Errorf("GetRatings returned empty result but no domain.ErrNotFoundError")
			}
			return fmt.Errorf("GetRatings : %v", err)
		}
		dishRatingOfUser = &ratings[0]
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

	dish, err := h.repo.GetDishByID(dbCtx, request.DishID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return PostDishesDishID404Response{}, nil
		}
		return PostDishesDishID500JSONResponse{}, nil
	}

	err = h.repo.CreateOrUpdateRating(dbCtx, userEmail, request.DishID,
		func(currentRating *domain.DishRating) (updatedRating *domain.DishRating, createNew bool, err error) {

			updatedRating = &dishRating
			createNew = dish.CreateNewRatingInsteadOfUpdating(currentRating, *updatedRating)
			err = nil
			return
		})
	if err != nil {
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
