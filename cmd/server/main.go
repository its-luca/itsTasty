package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"itsTasty/pkg/api/adapters/dishRepo"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports/botAPI"
	"itsTasty/pkg/api/ports/userAPI"
	"itsTasty/pkg/oidcAuth"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/errgroup"
)

const (
	envVarDBURL  = "DB_URL"
	envVarDBName = "DB_NAME"
	envVarDBUser = "DB_USER"
	envVarDBPW   = "DB_PW"

	envOIDCSecret                 = "OIDC_SECRET"
	envOIDCCallbackURL            = "OIDC_CALLBACK_URL"
	envOIDCProviderURL            = "OIDC_PROVIDER_URL"
	envOIDCID                     = "OIDC_ID"
	envOIDCRefreshIntervalMinutes = "OIDC_REFRESH_INTERVAL_MINUTES"

	envBotAPIToken = "BOT_API_TOKEN"

	//envURLAfterLogin sets the default url after login if no login target is given to the authenticator
	envURLAfterLogin  = "URL_AFTER_LOGIN"
	envURLAfterLogout = "URL_AFTER_LOGOUT"

	//envVarDevMode if set to true, mock login backend is used and X-Site cookies are allowed.
	//Make sure to also configure envVarDevCORS
	envVarDevMode = "DEV_MODE"

	//envVarDevCORS one URL that is allowed for CORS requests
	envVarDevCORS = "DEV_CORS"

	//envVarSessionLifetime is the maximum lifetime of the session cookie. Afterwards the user
	//has to log in again
	//see https://pkg.go.dev/time#ParseDuration for input format
	envVarSessionLifetime = "SESSION_LIFETIME"
)

type config struct {

	//DB config
	dbURL  string
	dbName string
	dbUser string
	dbPW   string

	//OIDC Config

	oidcSecret           string
	oidcCallbackURL      string
	oidcProviderURL      string
	oidcID               string
	urlAfterLogin        string
	urlAfterLogout       string
	oidcRefreshIntervall time.Duration

	//Bot Auth Config

	botAPIToken string

	// Session Config

	sessionSecret string

	//HTTP Config

	listen string

	//Config for local development

	devMode bool
	devCORS string

	//sessionLifetime is the expiry time of the session cookie
	sessionLifetime time.Duration
}

type application struct {
	conf          *config
	authenticator oidcAuth.Authenticator
	session       *scs.SessionManager
	router        chi.Router
	dishRepo      domain.DishRepo
	jobScheduler  *gocron.Scheduler
}

func parseConfig() (*config, error) {

	setEnvErr := func(envVarName string) error {
		return fmt.Errorf("missing env var %v", envVarName)
	}

	cfg := config{}

	cfg.devMode = "true" == strings.ToLower(os.Getenv(envVarDevMode))
	if cfg.devMode {
		log.Printf("DEV MODE: ENABLED")
	}

	if dbURL := os.Getenv(envVarDBURL); dbURL == "" {
		return nil, setEnvErr(envVarDBURL)
	} else {
		cfg.dbURL = dbURL
	}

	if dbName := os.Getenv(envVarDBName); dbName == "" {
		return nil, setEnvErr(envVarDBName)
	} else {
		cfg.dbName = dbName
	}

	if dbUser := os.Getenv(envVarDBUser); dbUser == "" {
		return nil, setEnvErr(envVarDBUser)
	} else {
		cfg.dbUser = dbUser
	}

	if dbPW := os.Getenv(envVarDBPW); dbPW == "" {
		return nil, setEnvErr(envVarDBPW)
	} else {
		cfg.dbPW = dbPW
	}

	if oidcSecret := os.Getenv(envOIDCSecret); oidcSecret == "" && !cfg.devMode {
		return nil, setEnvErr(envOIDCSecret)
	} else {
		cfg.oidcSecret = oidcSecret
	}

	if oidcCallbackURL := os.Getenv(envOIDCCallbackURL); oidcCallbackURL == "" && !cfg.devMode {
		return nil, setEnvErr(envOIDCCallbackURL)
	} else {
		cfg.oidcCallbackURL = oidcCallbackURL
	}

	if oidcProviderURL := os.Getenv(envOIDCProviderURL); oidcProviderURL == "" && !cfg.devMode {
		return nil, setEnvErr(envOIDCProviderURL)
	} else {
		cfg.oidcProviderURL = oidcProviderURL
	}

	if oidcID := os.Getenv(envOIDCID); oidcID == "" && !cfg.devMode {
		return nil, setEnvErr(envOIDCID)
	} else {
		cfg.oidcID = oidcID
	}

	{
		oidcRefreshIntervalStr := os.Getenv(envOIDCRefreshIntervalMinutes)
		//if not specified set default value
		if oidcRefreshIntervalStr == "" {
			cfg.oidcRefreshIntervall = 60 * time.Minute
			log.Printf("%v was not specified and defaults to %v", envOIDCRefreshIntervalMinutes, cfg.oidcRefreshIntervall)
		} else {
			oidcRefreshIntervalUint, err := strconv.ParseUint(oidcRefreshIntervalStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value %v of %s to uint64 : %v", oidcRefreshIntervalStr, envOIDCRefreshIntervalMinutes, err)
			}
			cfg.oidcRefreshIntervall = time.Duration(oidcRefreshIntervalUint) * time.Minute
		}

	}

	if botApiToken := os.Getenv(envBotAPIToken); botApiToken == "" {
		return nil, setEnvErr(envBotAPIToken)
	} else {
		cfg.botAPIToken = botApiToken
	}

	if urlAfterLogin := os.Getenv(envURLAfterLogin); urlAfterLogin == "" {
		return nil, setEnvErr(envURLAfterLogin)
	} else {
		cfg.urlAfterLogin = urlAfterLogin
	}

	if urlAfterLogout := os.Getenv(envURLAfterLogout); urlAfterLogout == "" {
		return nil, setEnvErr(envURLAfterLogout)
	} else {
		cfg.urlAfterLogout = urlAfterLogout
	}

	if devCORS := os.Getenv(envVarDevCORS); devCORS != "" {
		cfg.devCORS = devCORS
	}

	if sessionLifetimeAsStr := os.Getenv(envVarSessionLifetime); sessionLifetimeAsStr == "" {
		cfg.sessionLifetime = 7 * 24 * time.Hour
		log.Printf("%s was not specified and defaults to : %v", envVarSessionLifetime, cfg.sessionLifetime)
	} else {
		sessionLifetime, err := time.ParseDuration(sessionLifetimeAsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %v of %s to time duration : %v",
				sessionLifetimeAsStr, envVarSessionLifetime, err)
		}
		cfg.sessionLifetime = sessionLifetime
	}
	cfg.listen = ":80"

	return &cfg, nil
}

type dishRepoFactoryFunc func() (domain.DishRepo, error)

func newApplication(cfg *config, dishRepoFactory dishRepoFactoryFunc, botAPIFactory botAPI.ServiceFactory, userAPIFactory userAPI.HttpServerFactory) (*application, error) {

	//Build Session Storage

	log.Printf("Building session storage...")
	session := scs.New()
	session.Lifetime = cfg.sessionLifetime
	session.Cookie.Name = "its_tasty_session"
	session.Cookie.Secure = true

	if cfg.devMode {
		log.Printf("DEV MODE: Enabling SameSiteNoneMode")
		session.Cookie.SameSite = http.SameSiteNoneMode
	}

	//Build oidc authenticator
	log.Printf("Building auth backend...")
	authStorageAdapter := NewAuthSessionStorageManager(session)

	var authenticator oidcAuth.Authenticator
	if cfg.devMode {
		log.Printf("DEV MODE: Enabling Mock Login Backend")
		authenticator = oidcAuth.NewMockAuthenticator(cfg.urlAfterLogin, cfg.urlAfterLogout, authStorageAdapter)
	} else {
		var err error
		authenticator, err = oidcAuth.NewDefaultAuthenticator(cfg.oidcProviderURL, cfg.oidcID, cfg.oidcSecret,
			cfg.oidcCallbackURL, cfg.urlAfterLogin, cfg.urlAfterLogout, authStorageAdapter)
		if err != nil {
			return nil, fmt.Errorf("oidcAuth.NewDefaultAuthenticator : %v", err)
		}
	}

	//Schedule job to periodically refresh the oidc access tokens for all known sessions.
	//This way, our session token lifetime dictates the max lifetime of the session and not the oidc refresh
	//interval of the oidc provider

	jobScheduler := gocron.NewScheduler(time.Local)
	_, err := jobScheduler.Every(cfg.oidcRefreshIntervall).Do(func() {
		log.Printf("OIDC Refresh Job : Starting refresh...")
		//time limit for refresh jobs
		iterateCtx, iterateCtxCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer iterateCtxCancel()

		refreshJobs, _ := errgroup.WithContext(iterateCtx)
		refreshJobs.SetLimit(5)

		//number of sessions
		sessionCounter := 0
		//number of successfully refreshed sessions
		successCounter := 0
		successCounterLock := &sync.Mutex{}
		//iterateCtx over all active session and try to refresh them
		err := session.Iterate(iterateCtx, func(ctx context.Context) error {
			sessionCounter += 1
			//for logging purposes we want a reference to stick to our sessions, however printing the session
			//token itself would allow anyone with access to the logs to hijack that users session
			//thus we use the hash of the value isntead
			b := sha3.Sum512([]byte(session.Token(ctx)))
			tokenID := hex.EncodeToString(b[:])
			refreshJobs.Go(func() error {
				defer func() {
					_, _, err := session.Commit(ctx)
					if err != nil {
						log.Printf("OIDC Refresh Job: Token id %v, failed to commit session data :%v", tokenID, err)
						return
					}
				}()
				var err error
				for retries := 2; retries > 0; retries-- {
					err = authenticator.Refresh(ctx)
					if err == nil {
						break
					}
					log.Printf("OIDC Refresh Job: Refresh for token id %v failed (%v retries left) : %v", tokenID, retries, err)
					time.Sleep(3 * time.Second)
				}
				if err == nil {
					log.Printf("OIDC Refresh Job: Refreshed session  %v", tokenID)
					successCounterLock.Lock()
					successCounter += 1
					successCounterLock.Unlock()
					return nil
				}

				//if we get here, refresh has failed

				log.Printf("OIDC Refresh Job: Failed to refresh session %v : %v", tokenID, err)
				if err := session.Destroy(ctx); err != nil {
					log.Printf("OIDC Refresh Job: Failed to destroy session  %v after refresh failure : %v", tokenID, err)
				} else {
					log.Printf("OIDC Refresh Job: Destroyed session for %v after refresh failure : %v", tokenID, err)
				}

				//we do not want to propagate individual refresh errors beyond this function
				return nil
			})

			//very basic request throttling
			if sessionCounter%3 == 0 {
				time.Sleep(500 * time.Millisecond)
			}

			return nil
		})
		if err != nil {
			log.Printf("OIDC Refresh Job : Iterate failed : %v", err)
		}

		if err := refreshJobs.Wait(); err != nil {
			log.Printf("OIDC Refresh Job : errgroup.Wait failed : %v", err)
		}
		log.Printf("OIDC Refresh Job: Found %v sessions, refreshed %v, terminated %v", sessionCounter, successCounter, sessionCounter-successCounter)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to schedule session refresh job : %v", err)
	}

	repo, err := dishRepoFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate db backend : %v", err)
	}

	app := application{
		conf:          cfg,
		authenticator: authenticator,
		session:       session,
		dishRepo:      repo,
		jobScheduler:  jobScheduler,
	}

	app.router, err = app.setupRouter(botAPIFactory, userAPIFactory)
	if err != nil {
		return nil, fmt.Errorf("app.setupRouter : %v", err)
	}

	return &app, nil

}

// connectToMariaDB connects to the db waiting some time for the db to come online before giving up
func connectToPostgresDB(ctx context.Context, dbUser, dbPW, dbURL, dbName string) (*sql.DB, error) {

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@%s/%s?sslmode=disable", dbUser, dbPW, dbURL, dbName))
	if err != nil {
		return nil, fmt.Errorf("sql.Open : %v", err)
	}

	log.Printf("Trying to connect to db...")
	retries := 10
	connected := false
	for retries > 0 && !connected {
		pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
		err := db.PingContext(pingCtx)
		if err != nil {
			log.Printf("Error connecting to db : %v", err)
			log.Printf("%v retries remaining", retries)
			retries -= 1
			time.Sleep(3 * time.Second)
		} else {
			log.Printf("Connected to db")
			connected = true
		}
		pingCancel()
	}

	if !connected {
		return nil, fmt.Errorf("error connecting to db")
	}

	return db, nil
}

func (app *application) setupRouter(botAPIFactory botAPI.ServiceFactory, userAPiFactory userAPI.HttpServerFactory) (chi.Router, error) {
	log.Printf("Configuring router...")
	router := chi.NewRouter()

	if app.conf.devCORS != "" {
		log.Printf("DEV MODE: Allowing CORS + Credentials from %v", app.conf.devCORS)
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{app.conf.devCORS},
			AllowedMethods:   []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodHead, http.MethodPatch, http.MethodDelete},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			Debug:            false,
		}))
	}

	router.Use(app.session.LoadAndSave)
	router.Handle("/authAPI/callback", http.HandlerFunc(app.authenticator.CallbackHandler))
	router.Handle("/authAPI/login", http.HandlerFunc(app.authenticator.LoginHandler))
	router.Handle("/authAPI/logout", http.HandlerFunc(app.authenticator.LogoutHandler))
	//Builder User API for dishes

	userAPiRouter := chi.NewRouter()
	userAPiRouter.Use(app.authenticator.CheckSession)
	//only allow authenticated and setup context as api expects is
	userAPiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := oidcAuth.UserProfile{}
			raw := app.session.GetString(r.Context(), oidcAuth.SessionKeyProfile)
			if raw == "" {
				log.Printf("Blocked unauthenticated access to userAPI")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if err := json.Unmarshal([]byte(raw), &p); err != nil {
				log.Printf("Failed to unmarshal user profile : %v", err)
				http.Error(w, "", http.StatusInternalServerError)
			}

			//add user email to context
			r = r.WithContext(userAPI.ContextWithUserEmail(r.Context(), p.Email))

			next.ServeHTTP(w, r)
		})
	})

	userAPIServer := userAPiFactory(app.dishRepo)
	userAPIHandlers := userAPI.NewStrictHandler(userAPIServer, nil)
	userAPI.HandlerFromMux(userAPIHandlers, userAPiRouter)
	router.Mount("/userAPI/v1", userAPiRouter)

	//Build bot api
	botAPIRouter := chi.NewRouter()

	botAPIRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Bot APi api middleware called")
			gotAPIKey := r.Header.Get("X-API-KEY")
			log.Printf("parsed api key %v", gotAPIKey)
			if 0 == subtle.ConstantTimeCompare([]byte(gotAPIKey), []byte(app.conf.botAPIToken)) {
				log.Printf("wrong api key")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	botAPIServer := botAPIFactory(app.dishRepo)
	botAPIHandlers := botAPI.NewStrictHandler(botAPIServer, nil)
	botAPI.HandlerFromMux(botAPIHandlers, botAPIRouter)
	router.Mount("/botAPI/v1", botAPIRouter)

	//serve react frontend
	frontendRouter := chi.NewRouter()
	frontendRouter.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Printf("frontendRouter used for request to %v", request.URL.String())
			handler.ServeHTTP(writer, request)
		})
	})

	fsHandler := http.FileServer(http.Dir("/frontend"))
	frontendRouter.Handle("/", fsHandler)
	frontendRouter.Handle("/static/*", fsHandler)
	//see https://stackoverflow.com/questions/53876700/trying-to-serve-react-spa-that-uses-react-router
	frontendRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("not found hack for %v", r.URL.String())
		http.ServeFile(w, r, "/frontend/index.html")
	})
	router.Mount("/", frontendRouter)

	return router, nil
}

func run(app *application) error {
	mainCtx, mainCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer mainCancel()

	log.Printf("Starting http server on %v", app.conf.listen)
	srv := &http.Server{
		Handler: app.router,
		Addr:    app.conf.listen,
	}
	go func() {

		if !app.conf.devMode {
			if err := srv.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					log.Printf("Http server performs gracefull shutdown")
				} else {
					log.Printf("Error in HTTP server : %v", err)
				}
			}
		} else {
			const certPath = "./selfSignedTLS/server.crt"
			const keyPath = "./selfSignedTLS/server.key"
			log.Printf("DEV MODE: Starting with self signed TLS to be able to use SameSite=None. Make sure to"+
				"place self signed cert under %v and key under %v", certPath, keyPath)
			if err := srv.ListenAndServeTLS(certPath, keyPath); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					log.Printf("Http server performs gracefull shutdown")
				} else {
					log.Printf("Error in HTTP server : %v", err)
				}
			}
		}

	}()

	//start all jobs once at startup to quickly detect failures in otherwise seldom running jobs
	app.jobScheduler.RunAll()

	//run jobs periodically in background
	app.jobScheduler.StartAsync()

	log.Printf("Startup complete :)")
	<-mainCtx.Done()
	log.Printf("Beginning Shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()
	log.Printf("Shutting down http server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error requesting http server shutdown : %v", err)
	}

	return nil
}

func main() {

	log.SetFlags(log.Flags() | log.Llongfile)

	log.Printf("Parsing Config...")
	cfg, err := parseConfig()
	if err != nil {
		log.Printf("parseConfig : %v", err)
		return
	}

	defaultDishRepoFactory := func() (domain.DishRepo, error) {
		//Connect to db
		db, err := connectToPostgresDB(context.Background(), cfg.dbUser, cfg.dbPW, cfg.dbURL, cfg.dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect do db : %v", err)
		}

		//Build dish repo
		migrations := &migrate.FileMigrationSource{Dir: "/migrations/postgres"}
		repo, err := dishRepo.NewPostgresRepo(db, migrations)
		if err != nil {
			return nil, fmt.Errorf("failed to build mysql dish repo : %v", err)
		}
		return repo, nil
	}

	defaultBotApiFactory := func(repo domain.DishRepo) *botAPI.Service {
		return botAPI.NewService(repo)
	}

	defaultUserApiFactory := func(repo domain.DishRepo) *userAPI.HttpServer {
		return userAPI.NewHttpServer(repo)
	}

	log.Printf("Building application...")
	app, err := newApplication(cfg, defaultDishRepoFactory, defaultBotApiFactory, defaultUserApiFactory)
	if err != nil {
		log.Printf("newApplication : %v", err)
		return
	}

	log.Printf("Entering run...")
	if err := run(app); err != nil {
		log.Printf("run : %v", err)
		return
	}
}
