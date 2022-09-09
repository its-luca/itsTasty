#!/usr/bin/env bash

if [[ "$#" -ne 1 ]]; then
  echo "Usage: local-docker-build.sh <path to env file for frontend or 'dev' for a quick dev build>"
  exit 1
fi

ENV_FILE=$1

if [[ "$ENV_FILE" == 'dev' ]]; then
  sudo docker build -f ./cmd/server/Dockerfile -t "its-tasty:latest" .
  exit
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

sudo docker build \
  --build-arg  REACT_APP_USER_API_BASE_URL="$REACT_APP_USER_API_BASE_URL" \
  --build-arg  REACT_APP_AUTH_API_BASE_URL="$REACT_APP_AUTH_API_BASE_URL" \
  --build-arg  REACT_APP_PUBLIC_URL="$REACT_APP_PUBLIC_URL" \
  --build-arg  PUBLIC_URL="$PUBLIC_URL" \
  -f ./cmd/server/Dockerfile -t "its-tasty:latest" .