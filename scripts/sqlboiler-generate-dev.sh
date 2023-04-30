#!/bin/bash

source ./scripts/dev.env


export PSQL_HOST=localhost
export PSQL_DBNAME=${POSTGRES_DB}
export PSQL_USER=${POSTGRES_USER}
export PSQL_PASS=${POSTGRES_PASSWORD}
export PSQL_BLACKLIST="migrations"
export PSQL_SSLMODE="disable"
./scripts/sqlboiler-generate.sh