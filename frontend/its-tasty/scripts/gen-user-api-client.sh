#!/usr/bin/env bash

if [ "$#" -ne 1 ]; then
  echo "Usage gen-user-api-client.sh <path to openapi yml file>"
  exit 1
fi

API_SPEC=$1
./node_modules/.bin/openapi -i "$API_SPEC" -o src/services/userAPI

# Append code to make BASE URL configurable by env var
BASE_URL_ENV_VAR_NAME="REACT_APP_USER_API_BASE_URL"
printf '// @ts-ignore\nOpenAPI.BASE = process.env.%s\n' "$BASE_URL_ENV_VAR_NAME" >> src/services/userAPI/core/OpenAPI.ts
# Append code to set WITH_CREDENTIALS to true in dev mode. This sends our cookie in a cross site context
printf '// @ts-ignore\nOpenAPI.WITH_CREDENTIALS = process.env.NODE_ENV === "development" ? true : false\n'  >> src/services/userAPI/core/OpenAPI.ts


echo "You can configure the base url via the env var '$BASE_URL_ENV_VAR_NAME'"