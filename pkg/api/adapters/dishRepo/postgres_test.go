package dishRepo

import (
	"context"
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
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
	runCommonDbTests(t, func() (domain.DishRepo, factoryCleanupFunc, error) {
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
			err1 := db.Close()
			err2 := testutils.GlobalDockerPool.Cleanup()
			if err1 != nil || err2 != nil {
				return fmt.Errorf("db close err: %v ; cleanup err : %v", err1, err2)
			}
			return nil
		}
		return repo, cleanupFunc, nil
	})
}
