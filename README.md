# Messy deploy steps
1) Edit the `BASE` variable in `frontend/its-tasty/src/services/userAPI/core/OpenAPI.ts` to point to the correct URL (NO TRAILING SLASH) e.g. `https://docker.its.uni-luebeck.de/its-tasty/userAPI/v1`
2) Pay special attention to  `URL_AFTER_LOGIN`, `URL_AFTER_LOGOUT` and `OIDC_CALLBACK_URL` in the config, as these are absolute URL.
## Hosting in subdirectory
1) Edit `homepage` in `frontend/its-tasty/package.json` e.g. `https://docker.its.uni-luebeck.de/its-tasty`
2) Edit  the BrowserRouter's `basename` prop `frontend/its-tasty/src/index.tsx` e.g. `<BrowserRouter basename={"/its-tasty/"}>` to point to the sub path

