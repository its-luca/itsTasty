# Messy deploy steps
1) Edit the `BASE` variable in `frontend/its-tasty/src/services/userAPI/core/OpenAPI.ts` to point to the correct URL
(NO TRAILING SLASH) e.g. `https://docker.its.uni-luebeck.de/its-tasty/userAPI/v1`
2) Pay special attention to  `URL_AFTER_LOGIN`, `URL_AFTER_LOGOUT` and `OIDC_CALLBACK_URL` in the config,
as these are absolute URL.
3) For building the frontend, create an env file containing the variables defined in the `ALSO USED IN PRODUCTION`
section of `frontend/its-tasty/.env.development`.  Then use `frontend/its-tasty/scripts/build-release.sh <path to env file>`
to build the frontend. This supports subdirectory hosting!
