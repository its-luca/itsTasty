#!/usr/bin/env bash

DB_PORT=4896
CONTAINER_NAME="its-tasty-db-integration-test"


echo "Starting Mysql DB on ${DB_PORT}"
sudo docker run --rm --detach --name "${CONTAINER_NAME}" -p 4896:3306 \
  --env MARIADB_USER=integration_test_user \
  --env MARIADB_PASSWORD=1234 \
  --env MARIADB_ROOT_PASSWORD=rootpw \
  --env MARIADB_DATABASE=integration_test_db \
  mariadb:latest

#the test code uses these env vars to setup the db connection
export TEST_MYSQL_DB_LISTEN="localhost:${DB_PORT}"
export TEST_MYSQL_DB_USER="integration_test_user"
export TEST_MYSQL_DB_PW="1234"
export TEST_MYSQL_DB_NAME="integration_test_db"

echo "Running mysql integration tests..."
go clean -testcache
go test -v ./pkg/api/adapters/dishRepo

echo "Shutting down db"
sudo docker stop "${CONTAINER_NAME}"
