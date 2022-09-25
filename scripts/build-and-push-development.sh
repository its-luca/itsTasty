#!/usr/bin/env bash

TAG="development"

if [[ "$#" -ne 1 ]]; then
  echo "Usage: local-docker-build.sh <path to env file for frontend>"
  exit 1
fi

ENV_FILE=$1


DOCKER_TOKEN_FILE=./private/ghcr-token.txt
echo "$(cat $DOCKER_TOKEN_FILE)" | docker login  -u its-luca --password-stdin ghcr.io

# shellcheck disable=SC1090
source "$ENV_FILE"

sudo docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg  REACT_APP_USER_API_BASE_URL="$REACT_APP_USER_API_BASE_URL" \
  --build-arg  REACT_APP_AUTH_API_BASE_URL="$REACT_APP_AUTH_API_BASE_URL" \
  --build-arg  REACT_APP_PUBLIC_URL="$REACT_APP_PUBLIC_URL" \
  --build-arg  PUBLIC_URL="$PUBLIC_URL" \
  -f cmd/server/Dockerfile --push -t "ghcr.io/its-luca/itstasty:$TAG" .

