#!/usr/bin/env bash

if [ "$#" -ne 1 ]; then
  echo "Usage gen-user-api-client.sh <path to openapi yml file"
  exit
fi

API_SPEC=$1
./node_modules/.bin/openapi -i "$API_SPEC" -o src/services/userAPI

echo "Dont forget to adjust the config in openapi/core/OpenAPI.ts"