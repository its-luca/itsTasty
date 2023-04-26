package userAPI

import (
	"context"
	"errors"
	"fmt"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports"
	"log"
	"time"

	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
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

const defaultDBTimeout = 5 * time.Second

type HttpServer struct {
	repo       domain.DishRepo
	timeSource TimeSource
}

func (h *HttpServer) GetDishesMergeCandidatesDishID(ctx context.Context, request GetDishesMergeCandidatesDishIDRequestObject) (GetDishesMergeCandidatesDishIDResponseObject, error) {

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout*2)
	defer dbCancel()

	mergeCandidates, err := ports.FetchMergeCandidates(dbCtx, request.DishID, h.repo)
	if err != nil {
		log.Printf("ports.FetchMergeCandidates failed with : %v", err)
		if errors.Is(err, domain.ErrNotFound) {
			return GetDishesMergeCandidatesDishID404Response{}, nil
		}
		return GetDishesMergeCandidatesDishID500Response{}, nil
	}

	respData := make([]GetMergeCandidatesRespEntry, 0)
	for i := range mergeCandidates {
		v := &mergeCandidates[i]

		if v.SimilarityScore >= ports.MergeCandidatesDefaultSimilarityThresh {
			respData = append(respData, GetMergeCandidatesRespEntry{
				DishID:       v.DishID,
				DishName:     v.Name,
				MergedDishID: v.MergedDishID,
			})
		}
	}

	return GetDishesMergeCandidatesDishID200JSONResponse{Candidates: respData}, nil

}

func NewHttpServer(repo domain.DishRepo) *HttpServer {
	return &HttpServer{repo: repo, timeSource: defaultTimeSource{}}
}

type HttpServerFactory func(repo domain.DishRepo) *HttpServer

func NewHttpServerCustomTime(repo domain.DishRepo, timeSource TimeSource) *HttpServer {
	return &HttpServer{
		repo:       repo,
		timeSource: timeSource,
	}
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
	if len(dishNames) != len(dishIDs) {
		log.Printf("Got %v dishNames but only %v dishIds", len(dishNames), len(dishIDs))
		return GetMergedDishesMergedDishID500Response{}, nil
	}
	containedDishes := make([]ContainedDishEntry, 0)
	for idx, id := range dishIDs {
		containedDishes = append(containedDishes, ContainedDishEntry{
			Id:   id,
			Name: dishNames[idx],
		})
	}

	//assemble response

	resp := GetMergedDishesMergedDishID200JSONResponse{
		Name:            mergedDish.Name,
		ServedAt:        mergedDish.ServedAt,
		ContainedDishes: containedDishes,
	}

	return resp, nil
}

// userFacingDishNotExistsErr helper err to generate error messages for NotFound/400 http error
type userFacingDishNotExistsErr struct {
	dishID int64
}

func (u userFacingDishNotExistsErr) Error() string {
	return fmt.Sprintf("Dish with id %v does not exist", u.dishID)
}

func (h *HttpServer) PostMergedDishes(ctx context.Context, request PostMergedDishesRequestObject) (PostMergedDishesResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	if len(request.Body.MergedDishes) < 2 {
		s := "You must provide at least two dish IDs"
		return PostMergedDishes400JSONResponse{What: &s}, nil
	}

	//
	// Fetch Resources
	//

	mapper := iter.Mapper[int64, *domain.Dish]{}
	dishesForMerge, err := mapper.MapErr(request.Body.MergedDishes, func(i *int64) (*domain.Dish, error) {
		d, err := h.repo.GetDishByID(dbCtx, *i)
		if err != nil {
			log.Printf("GetDishByID for id %v failed : %v", *i, err)
			if errors.Is(err, domain.ErrNotFound) {

				return nil, userFacingDishNotExistsErr{*i}
			}
			return nil, fmt.Errorf("GetDishByID failed : %v", err)
		}
		return d, nil
	})
	if err != nil {
		var notFoundErr userFacingDishNotExistsErr
		if errors.As(err, &notFoundErr) {
			s := notFoundErr.Error()
			return PostMergedDishes400JSONResponse{What: &s}, nil
		}
		return PostMergedDishes500Response{}, nil
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

	if request.Body.Name != nil && *request.Body.Name == "" {
		what := "Name may not be empty string"
		return PatchMergedDishesMergedDishID400JSONResponse{What: &what}, nil
	}

	mapper := iter.Mapper[int64, *domain.Dish]{}
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

		if request.Body.Name != nil {
			current.Name = *request.Body.Name
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

func fetchMostRecentUserRating(ctx context.Context, repo domain.DishRepo, userEmail string, dishID int64) (*domain.DishRating, error) {
	ratings, err := repo.GetRatings(ctx, userEmail, dishID, true)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("GetRatings : %v", err)
	}

	return &ratings[0], nil
}

func (h *HttpServer) GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) (GetDishesDishIDResponseObject, error) {
	userEmail, err := GetUserEmailFromCTX(ctx)
	if err != nil {
		log.Printf("GetUserEmailFromCTX : %v", err)
		return GetDishesDishID500JSONResponse{}, nil
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	p := pool.New().WithContext(dbCtx).WithCancelOnError()

	var basicDishData *ports.BasicDishReply
	var mostRecentUserRating *domain.DishRating

	p.Go(func(ctx context.Context) error {
		var err error
		basicDishData, err = ports.FetchBasicDishData(dbCtx, h.repo, request.DishID)
		return err
	})
	p.Go(func(ctx context.Context) error {
		var err error
		mostRecentUserRating, err = fetchMostRecentUserRating(dbCtx, h.repo, userEmail, request.DishID)
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	})

	if err := p.Wait(); err != nil {
		log.Printf("failed to fetch data for reply : %v", err)
		return GetDishesDishID500JSONResponse{}, nil
	}

	response := GetDishesDishID200JSONResponse{
		AvgRating:         basicDishData.AvgRating, //updated below if data is available
		Name:              basicDishData.Name,
		ServedAt:          basicDishData.ServedAt,
		OccurrenceCount:   basicDishData.OccurrenceCount,
		RatingOfUser:      nil, //updated below if data is available
		Ratings:           basicDishData.Ratings,
		RecentOccurrences: basicDishData.RecentOccurrences,
		MergedDishID:      basicDishData.MergedDishID,
	}

	if mostRecentUserRating != nil {
		v := GetDishRespRatingOfUser(mostRecentUserRating.Value)
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
		Who:        userEmail,
		Value:      rating,
		RatingWhen: h.timeSource.Now(),
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
	dishesSimple, err := h.repo.GetAllDishesSimple(dbCtx)
	if err != nil {
		log.Printf("GetAllDishes : %v", err)
		return GetGetAllDishes500JSONResponse{}, nil
	}

	respData := make([]GetAllDishesRespEntry, 0, len(dishesSimple))
	for _, v := range dishesSimple {
		respData = append(respData, GetAllDishesRespEntry{
			Id:           v.Id,
			MergedDishID: v.MergedDishID,
			Name:         v.Name,
			ServedAt:     v.ServedAt,
		})
	}

	return GetGetAllDishes200JSONResponse{
		Data: respData,
	}, nil
}

func (h *HttpServer) PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) (PostSearchDishResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	//Search both merged dishes and dishes for the given name + location combination
	//For a merged dish, return the id of the most recently served dish
	//Prefer hits on merged dishes, as they are allowed to shadow their contained dishes by having the same
	//name as one of them

	p := pool.New().WithContext(dbCtx)

	var dishIDFromMergedDish *int64
	p.Go(func(ctx context.Context) error {
		_, mergedDishID, err := h.repo.GetMergedDish(ctx, request.Body.DishName, request.Body.ServedAt)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			return err
		}
		_, idMostRecentDish, err := h.repo.GetMostRecentDishForMergedDish(ctx, mergedDishID)
		if err != nil {
			return fmt.Errorf("found matching merged dish but failed to get most recently served dish : %v", err)
		}
		*dishIDFromMergedDish = idMostRecentDish
		return nil
	})

	var dishID *int64
	p.Go(func(ctx context.Context) error {
		_, id, err := h.repo.GetDishByName(dbCtx, request.Body.DishName, request.Body.ServedAt)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			return err
		}
		*dishID = id
		return nil
	})

	if err := p.Wait(); err != nil {
		log.Printf("GetDishByName for %v: %v", request.Body.DishName, err)
		return PostSearchDish500JSONResponse{}, nil
	}

	if dishIDFromMergedDish == nil && dishID == nil {
		return PostSearchDish200JSONResponse{
			DishName:  request.Body.DishName,
			FoundDish: false,
		}, nil
	}

	var resultDishID *int64
	if dishIDFromMergedDish != nil {
		resultDishID = dishIDFromMergedDish
	} else {
		resultDishID = dishID
	}

	return PostSearchDish200JSONResponse{
		DishID:    resultDishID,
		DishName:  request.Body.DishName,
		FoundDish: true,
	}, nil
}
