package ports

import (
	"context"
	"fmt"
	"itsTasty/pkg/api/domain"
	"log"
	"sort"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type MergeCandidate struct {
	DishID               int64
	Name                 string
	SimilarityScore      float64
	MergedDishID         *int64
	PreprocessedBaseName string
	PreprocessedName     string
}

const MergeCandidatesDefaultSimilarityThresh = 0.55

// FetchMergeCandidates generates a slice of all potential merge candidate for baseDishID
// Marker errors domain.ErrNotFound if baseDishID is not found
func FetchMergeCandidates(ctx context.Context, baseDishID int64, repo domain.DishRepo) ([]MergeCandidate, error) {

	//
	// query requested dish + ids of all other dishes
	//
	baseDish, err := repo.GetDishByID(ctx, baseDishID)
	if err != nil {
		log.Printf("Failed to fetch dish %v : %v", baseDishID, err)

		return nil, fmt.Errorf("GetDishByID failed : %w", err)
	}

	allDishesSimple, err := repo.GetAllDishesSimple(ctx)
	if err != nil {
		log.Printf("Failed to fetch all dish ids : %v", err)

		return nil, fmt.Errorf("GetAllDishesSimple failed : %w", err)
	}

	mergeCandidates := make([]MergeCandidate, 0, len(allDishesSimple))
	for _, candidateDish := range allDishesSimple {

		//if served in a different location, we never want to merge
		//Future versions should filter this in the db query
		if candidateDish.ServedAt != baseDish.ServedAt {
			mergeCandidates = append(mergeCandidates, MergeCandidate{
				DishID:          candidateDish.Id,
				Name:            candidateDish.Name,
				SimilarityScore: 0,
			})
			continue
		}

		//we don't want to merge a candidateDish with itself
		if candidateDish.Id == baseDishID {
			mergeCandidates = append(mergeCandidates, MergeCandidate{
				DishID:          candidateDish.Id,
				Name:            candidateDish.Name,
				SimilarityScore: 0,
			})
			continue
		}

		//remove stop words
		candidateNamePreprocessed := strings.ToLower(candidateDish.Name)
		baseNamePreprocessed := strings.ToLower(baseDish.Name)
		for _, stopWord := range []string{"auf", "mit", "dazu", "und", "-", ",", "auch als kleine Portion"} {
			candidateNamePreprocessed = strings.ReplaceAll(candidateNamePreprocessed, stopWord, " ")
			baseNamePreprocessed = strings.ReplaceAll(baseNamePreprocessed, stopWord, " ")
		}
		tokeniseSortReasemble := func(s string) string {
			tokens := strings.Split(s, " ")
			sort.Slice(tokens, func(i, j int) bool {
				return tokens[i] < tokens[j]
			})
			return strings.Join(tokens, " ")
		}
		candidateNamePreprocessed = tokeniseSortReasemble(candidateNamePreprocessed)
		baseNamePreprocessed = tokeniseSortReasemble(baseNamePreprocessed)

		//return levenshtein as similarity score
		similarity := strutil.Similarity(candidateNamePreprocessed, baseNamePreprocessed, metrics.NewLevenshtein())

		mergeCandidates = append(mergeCandidates, MergeCandidate{
			DishID:               candidateDish.Id,
			Name:                 candidateDish.Name,
			SimilarityScore:      similarity,
			MergedDishID:         candidateDish.MergedDishID,
			PreprocessedBaseName: baseNamePreprocessed,
			PreprocessedName:     candidateNamePreprocessed,
		})
	}

	return mergeCandidates, nil
}
