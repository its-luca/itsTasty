package main

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"log"
	"sync"
)

type dockerPool struct {
	lock         *sync.Mutex
	pool         *dockertest.Pool
	resource     *dockertest.Resource
	usageCounter int
}

var globalDockerPool = &dockerPool{
	lock:         &sync.Mutex{},
	pool:         nil,
	resource:     nil,
	usageCounter: 0,
}

func getPostgresIntegrationTestDB(config *dockerPool) (*sql.DB, error) {
	const pgUser = "its_tasty_user"
	const pgPW = "12345"

	config.lock.Lock()
	defer func() {
		config.lock.Unlock()
	}()

	if config.pool == nil {
		log.Printf("Initializing docker pool...")
		pool, err := dockertest.NewPool("")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to docker : %v", err)
		}
		config.pool = pool
		config.usageCounter = 0
		log.Printf("Pool initializaiton done")
	}

	if config.usageCounter == 0 {
		log.Printf("Creating container...")
		//resource, err := config.pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=" + pgPW, "POSTGRES_DB=" + testDBName, "POSTGRES_USER=" + pgUser})
		resource, err := config.pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=" + pgPW, "POSTGRES_USER=" + pgUser})
		if err != nil {
			return nil, fmt.Errorf("failed to start postgres controlDB : %v", err)
		}
		if err := resource.Expire(5 * 60); err != nil {
			return nil, fmt.Errorf("failed to set hard tiemout for container : %v", err)
		}

		config.resource = resource
		log.Printf("Container Initialization done")

	}
	log.Printf("Waiting for postgres to accept connections...")
	var db *sql.DB
	err := config.pool.Retry(func() error {
		var err error
		controlDB, err := sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@localhost:%s?sslmode=disable", pgUser, pgPW, config.resource.GetPort("5432/tcp")))
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
		dbName := fmt.Sprintf("test_db_%v", config.usageCounter)
		if _, err := controlDB.Exec("CREATE DATABASE " + dbName); err != nil {
			return fmt.Errorf("failed to create database %v : %v", dbName, err)
		}

		db, err = sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@localhost:%s/%s?sslmode=disable", pgUser, pgPW, config.resource.GetPort("5432/tcp"), dbName))
		if err != nil {
			return fmt.Errorf("failed to connect to db : %v", err)
		}
		return nil
	})
	if err != nil {
		log.Printf("retry for connecting to controlDB timed out : %v", err)
	}

	config.usageCounter += 1

	return db, nil
}

func cleanup(config *dockerPool) error {
	config.lock.Lock()
	defer func() {
		config.lock.Unlock()
	}()

	config.usageCounter -= 1

	if config.usageCounter > 0 {
		return nil
	}

	//no more usage : clean up container
	if err := config.pool.Purge(config.resource); err != nil {
		return fmt.Errorf("failed to shutdown container : %v", err)
	}
	return nil
}
