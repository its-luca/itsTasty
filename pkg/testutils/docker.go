package testutils

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"log"
	"sync"
)

type dockerPool struct {
	lock          *sync.Mutex
	pool          *dockertest.Pool
	resource      *dockertest.Resource
	usageCounter  int
	uniqueCounter int
}

var GlobalDockerPool = &dockerPool{
	lock:          &sync.Mutex{},
	pool:          nil,
	resource:      nil,
	usageCounter:  0,
	uniqueCounter: 0,
}

// uniqueNumberLocked is a helper that returns a unique number
// MUST be called under dockerPool.lock !!!
func (p *dockerPool) uniqueNumberLocked() int {
	v := p.uniqueCounter
	p.uniqueCounter += 1
	return v
}

// GetPostgresIntegrationTestDB creates a DB handle using a new, unique postgres database
// You must call Cleanup once you are done using the db
func (p *dockerPool) GetPostgresIntegrationTestDB() (*sql.DB, error) {
	const pgUser = "its_tasty_user"
	const pgPW = "12345"

	p.lock.Lock()
	defer func() {
		p.lock.Unlock()
	}()

	if p.pool == nil {
		log.Printf("Initializing docker pool...")
		pool, err := dockertest.NewPool("")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to docker : %v", err)
		}
		p.pool = pool
		p.usageCounter = 0
		log.Printf("Pool initializaiton done")
	}

	if p.usageCounter == 0 {
		log.Printf("Creating container...")
		//resource, err := p.pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=" + pgPW, "POSTGRES_DB=" + testDBName, "POSTGRES_USER=" + pgUser})
		resource, err := p.pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=" + pgPW, "POSTGRES_USER=" + pgUser})
		if err != nil {
			return nil, fmt.Errorf("failed to start postgres controlDB : %v", err)
		}
		if err := resource.Expire(5 * 60); err != nil {
			return nil, fmt.Errorf("failed to set hard tiemout for container : %v", err)
		}

		p.resource = resource
		log.Printf("Container Initialization done")

	}
	log.Printf("Waiting for postgres to accept connections...")
	var db *sql.DB
	err := p.pool.Retry(func() error {
		var err error
		controlDB, err := sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@localhost:%s?sslmode=disable", pgUser, pgPW, p.resource.GetPort("5432/tcp")))
		if err != nil {
			return fmt.Errorf("failed to connect to controlDB : %v", err)
		}
		defer func() {
			_ = controlDB.Close()
		}()
		if err := controlDB.Ping(); err != nil {
			return fmt.Errorf("ping failed : %v", err)
		}
		//create new unique controlDB and switch to it before returning the controlDB handle
		dbName := fmt.Sprintf("test_db_%v", p.uniqueNumberLocked())
		if _, err := controlDB.Exec("CREATE DATABASE " + dbName); err != nil {
			return fmt.Errorf("failed to create database %v : %v", dbName, err)
		}

		db, err = sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@localhost:%s/%s?sslmode=disable", pgUser, pgPW, p.resource.GetPort("5432/tcp"), dbName))
		if err != nil {
			return fmt.Errorf("failed to connect to db : %v", err)
		}
		return nil
	})
	if err != nil {
		log.Printf("retry for connecting to controlDB timed out : %v", err)
	}

	p.usageCounter += 1

	return db, nil
}

// Cleanup must be called once the previously requested container backed instance is no longer used.
// It is used to infer when we can shut down the container
func (p *dockerPool) Cleanup() error {
	p.lock.Lock()
	defer func() {
		p.lock.Unlock()
	}()

	p.usageCounter -= 1

	if p.usageCounter > 0 {
		return nil
	}

	//no more usage : clean up container
	if err := p.pool.Purge(p.resource); err != nil {
		return fmt.Errorf("failed to shutdown container : %v", err)
	}
	return nil
}
