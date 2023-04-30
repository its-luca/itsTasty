package dishRepo

import (
	"context"
	"errors"
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/testutils"
	"testing"
)

func Test_NewPostgresRepo(t *testing.T) {
	db, err := testutils.GlobalDockerPool.GetPostgresIntegrationTestDB()
	if err != nil {
		t.Fatalf("GetPostgresIntegrationTestDB failed : %v", err)
	}
	defer func() {
		if err := testutils.GlobalDockerPool.Cleanup(); err != nil {
			t.Fatalf("failed to cleanup docker pool : %v", err)
		}
	}()

	migrationSource := &migrate.FileMigrationSource{Dir: "../../../../migrations/postgres"}
	repo, err := NewPostgresRepo(db, migrationSource)
	require.NoError(t, err)

	err = repo.DropRepo(context.Background())
	require.NoError(t, err)
}

func Test_Postgres_RunCommon(t *testing.T) {

	commonFactory := func() (*PostgresRepo, factoryCleanupFunc, error) {
		db, err := testutils.GlobalDockerPool.GetPostgresIntegrationTestDB()
		if err != nil {
			t.Fatalf("GetPostgresIntegrationTestDB failed : %v", err)
		}
		migrationSource := &migrate.FileMigrationSource{Dir: "../../../../migrations/postgres"}
		repo, err := NewPostgresRepo(db, migrationSource)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create repo : %v", err)
		}

		cleanupFunc := func() error {
			err1 := repo.DropRepo(context.Background())
			err2 := db.Close()
			err3 := testutils.GlobalDockerPool.Cleanup()
			err := errors.Join(err1, err2, err3)
			if err != nil {
				return fmt.Errorf("cleanupFunc failed : %v", err)
			}
			return nil
		}
		return repo, cleanupFunc, nil
	}

	dishFactory := func() (domain.DishRepo, factoryCleanupFunc, error) {
		return commonFactory()
	}
	streakFactory := func() (domain.RatingStreakRepo, factoryCleanupFunc, error) {
		return commonFactory()
	}

	runCommonDbTests(t, dishFactory, streakFactory, commonFactory)
}

func Test_arrayDiff(t *testing.T) {
	type args[K comparable] struct {
		old     []K
		updated []K
	}
	type testCase[K comparable] struct {
		name        string
		args        args[K]
		wantRemoved []K
		wantAdded   []K
	}
	tests := []testCase[int]{
		{
			name: "only added values",
			args: args[int]{
				old:     []int{1, 2, 3},
				updated: []int{1, 2, 3, 4, 5, 6},
			},
			wantRemoved: []int{},
			wantAdded:   []int{4, 5, 6},
		},
		{
			name: "only removed values",
			args: args[int]{
				old:     []int{1, 2, 3},
				updated: []int{2},
			},
			wantRemoved: []int{1, 3},
			wantAdded:   []int{},
		},
		{
			name: "only unchanged",
			args: args[int]{
				old:     []int{1, 2, 3},
				updated: []int{1, 2, 3},
			},
			wantRemoved: []int{},
			wantAdded:   []int{},
		},
		{
			name: "old is empty",
			args: args[int]{
				old:     []int{},
				updated: []int{1, 2, 3},
			},
			wantRemoved: []int{},
			wantAdded:   []int{1, 2, 3},
		},
		{
			name: "updated is empty",
			args: args[int]{
				old:     []int{1},
				updated: []int{},
			},
			wantRemoved: []int{1},
			wantAdded:   []int{},
		},
		{
			name: "removed, updated, same combined",
			args: args[int]{
				old:     []int{1, 2},
				updated: []int{2, 3},
			},
			wantRemoved: []int{1},
			wantAdded:   []int{3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemoved, gotAdded := arrayDiff(tt.args.old, tt.args.updated)
			assert.ElementsMatchf(t, tt.wantRemoved, gotRemoved, "unexpected deleted values")
			assert.ElementsMatchf(t, tt.wantAdded, gotAdded, "unexpected added values")

		})
	}
}
