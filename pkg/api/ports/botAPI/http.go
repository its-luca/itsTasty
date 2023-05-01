package botAPI

import (
	"context"
	"errors"
	"fmt"
	"github.com/sourcegraph/conc/pool"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports"
	"itsTasty/pkg/api/statisticsService"
	"log"
	"strings"
	"time"
)

//go:generate oapi-codegen --config ./server.cfg.yml ../../botAPI.yml
//go:generate oapi-codegen --config ./types.cfg.yml ../../botAPI.yml

const defaultDBTimeout = 5 * time.Second

type Service struct {
	repo          domain.DishRepo
	streakService statisticsService.StreakService
	timeSource    TimeSource
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

func NewService(repo domain.DishRepo, streak statisticsService.StreakService) *Service {
	return &Service{
		repo:          repo,
		streakService: streak,
		timeSource:    defaultTimeSource{},
	}
}

type ServiceFactory func(repo domain.DishRepo, streakService statisticsService.StreakService) *Service

func NewServiceCustomTime(repo domain.DishRepo, streakService statisticsService.StreakService, timeSource TimeSource) *Service {
	return &Service{
		repo:          repo,
		streakService: streakService,
		timeSource:    timeSource,
	}
}

func (s *Service) GetStatisticsCurrentVotingStreaks(ctx context.Context, request GetStatisticsCurrentVotingStreaksRequestObject) (GetStatisticsCurrentVotingStreaksResponseObject, error) {

	//try to update rating streaks before processing response. Skip this step if it takes to long
	//(vacation backend queried by statistics service is known to be unreliable)
	updateCtx, updateCancel := context.WithTimeout(ctx, 10*time.Second)
	defer updateCancel()
	if err := s.streakService.UpdateRatingStreaks(updateCtx); err != nil {
		log.Printf("Failed to update rating streaks : %v", err)
		if !errors.Is(err, context.DeadlineExceeded) {
			return GetStatisticsCurrentVotingStreaks500Response{}, nil
		}
	}

	//fetch response data
	fetchPool := pool.New().WithContext(ctx).WithCancelOnError()

	usersWithLongestStreak := make([]string, 0)
	var longestUserStreak *int
	fetchPool.Go(func(ctx context.Context) error {
		usersWithStreak, err := s.streakService.GetMostRecentUserStreaks(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch user streaks : %v", err)
		}
		if len(usersWithStreak) == 0 {
			return nil
		}

		//determine user(s) with longest streak and store in response format
		usersWithLongestStreak = []string{usersWithStreak[0].User.Email}
		tmp := usersWithStreak[0].MostRecentStreak.LengthInDays()
		longestUserStreak = &tmp
		for _, v := range usersWithStreak[1:] {
			if v.MostRecentStreak.LengthInDays() == *longestUserStreak {
				usersWithLongestStreak = append(usersWithLongestStreak, v.User.Email)
			} else if v.MostRecentStreak.LengthInDays() > *longestUserStreak {
				tmp = v.MostRecentStreak.LengthInDays()
				longestUserStreak = &tmp
				usersWithLongestStreak = []string{v.User.Email}
			}

		}
		return nil
	})

	var allUsersStreakLength *int
	fetchPool.Go(func(ctx context.Context) error {
		teamStreak, err := s.streakService.GetMostRecentAllUsersGroupStreak(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("failed to fetch \"all users\" streak : %v", err)

		}

		tmp := teamStreak.LengthInDays()
		allUsersStreakLength = &tmp
		return nil
	})

	if err := fetchPool.Wait(); err != nil {
		log.Printf("at least one request in fetchPool failed : %v", err)
		return GetStatisticsCurrentVotingStreaks500Response{}, nil
	}

	//assemble response
	response := GetStatisticsCurrentVotingStreaks200JSONResponse{
		CurrentTeamVotingStreak:       allUsersStreakLength,
		CurrentUserVotingStreakLength: longestUserStreak,
		UsersWithMaxStreak: func() *[]string {
			if len(usersWithLongestStreak) == 0 {
				return nil
			} else {
				return &usersWithLongestStreak
			}
		}(),
	}

	return response, nil
}

func (s *Service) PostCreateOrUpdateDish(ctx context.Context, request PostCreateOrUpdateDishRequestObject) (PostCreateOrUpdateDishResponseObject, error) {
	dbCtx, dbCancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer dbCancel()

	sanitizeName := func(s string) string {
		prefixes := []string{
			`"""YOUR FAVORITES""`,
			`"YOUR FAVORITES"`,
			"Begrenztes Angebot :",
			"BEGRENZTES ANGEBOT:",
			"VEGANISSIMO: ",
		}
		for _, v := range prefixes {
			s = strings.TrimPrefix(s, v)
		}

		s = strings.Trim(s, " ")
		return s

	}

	request.Body.DishName = sanitizeName(request.Body.DishName)

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

	checkMergeCandidates := false
	//set checkMergeCandidates to true if we have at least one
	if createdDish {
		dbCancel()
		dbCtx, dbCancel = context.WithTimeout(ctx, defaultDBTimeout)
		defer dbCancel()
		mergeCandidates, err := ports.FetchMergeCandidates(dbCtx, dishID, s.repo)
		if err != nil {
			log.Printf("ports.FetchMergeCandidates failed with : %v", err)
			if errors.Is(err, domain.ErrNotFound) {
				log.Printf("domain.ErrNotFound should never happen here, since we just created the dish")
			}
			return PostCreateOrUpdateDish500JSONResponse{}, nil
		}

		for i := range mergeCandidates {
			v := &mergeCandidates[i]

			if v.SimilarityScore >= ports.MergeCandidatesDefaultSimilarityThresh {
				checkMergeCandidates = true
				break
			}
		}
	}

	//
	// Assemble Response
	//

	return PostCreateOrUpdateDish200JSONResponse{
		CreatedNewDish:       createdDish,
		CreatedNewLocation:   createdLocation,
		DishID:               dishID,
		CheckMergeCandidates: checkMergeCandidates,
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
