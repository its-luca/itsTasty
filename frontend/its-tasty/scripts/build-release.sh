#!/usr/bin/env bash

# This script loads and env file before issuing `npm run build` to allow "docker-style" builds for different env
# in a more explicit manner than the magic name env files supported by React

if [ "$#" -ne 1 ]; then
  echo "Usage build-release.sh <path to env file"
  exit 1
fi

ENV_FILE=$1

#load and export all vars in ENV_FILE; see https://unix.stackexchange.com/questions/79064/how-to-export-variables-from-a-file
set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a
npm run build