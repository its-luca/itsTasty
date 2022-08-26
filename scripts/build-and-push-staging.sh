#!/usr/bin/env bash

TAG="staging"
#if [ "$1" != "" ]; then
#  TAG=$1
#fi

DOCKER_TOKEN_FILE=./private/ghcr-token.txt
echo "$(cat $DOCKER_TOKEN_FILE)" | docker login  -u its-luca --password-stdin ghcr.io



sudo docker buildx build --platform linux/amd64,linux/arm64 -f cmd/server/Dockerfile --push -t "ghcr.io/its-luca/itstasty/its-tasty:$TAG" .

