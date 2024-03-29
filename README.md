# Manual Deploy
This app is deployed via docker. We have two kind of environment files that need to be configured. 

**At build time**, we need an environment file for configuring things like the PUBLIC_URL etc. for our React app.
See the `ALSO USED IN PRODUCTION` section in `./frontend/its-tasty/.env.development` for a template. The path to the 
deployment env file needs to be passed to the docker build script in the `./scripts` folder

**When starting the container** you need to pass an env file as usual.


# Development/Debugging
In Production, we serve both the backend and the frontend from the same webserver allowing us to use regular cookies
for login. Also, we need access to an OIDC Server for the login.
For development, we serve the backend and the frontend on two different ports on localhost, requiring us to enable
cookies in a cross site context. Furthermore, the OIDC dependency is replaced with a mock that can login and logout a fixed
user.
Start the backend with `./scripts/dev.env` env file. To enable this setup.
Furthermore, you need to generate a local TLS cert once with `./scripts/gen-local-certs.sh`  as enable cross site
cookies requires serving via TLS.

The frontend should automatically pick up its config from `frontend/its-tasty/.env.development`

## Generate SQL Code
This repo users [sqlboiler](https://github.com/volatiletech/sqlboiler) to manage sql boilerplate code and
[sql-migrate](https://github.com/rubenv/sql-migrate) to manage db migrations.
To change sth. in the db
1) Spin up a docker container with postgres using the credentials from `./scripts/dev.env`. You can use `its-tasty-db`
from the `docker-compose.yml`. Just make sure to erase the db volume to get a fresh state.
2) Run `./scripts/sql-migrate-dev-db.sh`
3) Run `./scripts/sqlboiler-generate-dev.sh`