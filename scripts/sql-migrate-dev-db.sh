#!/bin/bash

source ./scripts/dev.env

export PSQL_DBNAME=${POSTGRES_DB}
export PSQL_USER=${POSTGRES_USER}
export PSQL_PASSWORD=${POSTGRES_PASSWORD}
export sslmode=disable

sql-migrate  up