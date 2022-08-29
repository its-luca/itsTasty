package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"itsTasty/pkg/api/adapters/dishRepo"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports/botAPI"
	"itsTasty/pkg/api/ports/userAPI"
	"itsTasty/pkg/oidcAuth"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envVarDBURL  = "DB_URL"
	envVarDBName = "DB_NAME"
	envVarDBUser = "DB_USER"
	envVarDBPW   = "DB_PW"

	envOIDCSecret      = "OIDC_SECRET"
	envOIDCCallbackURL = "OIDC_CALLBACK_URL"
	envOIDCProviderURL = "OIDC_PROVIDER_URL"
	envOIDCID          = "OIDC_ID"

	envBotAPIToken = "BOT_API_TOKEN"
)

type config struct {

	//DB config
	dbURL  string
	dbName string
	dbUser string
	dbPW   string

	//OIDC Config

	oidcSecret      string
	oidcCallbackURL string
	oidcProviderURL string
	oidcID          string

	//Bot Auth Config

	botAPIToken string

	// Session Config

	sessionSecret string

	//HTTP Config

	listen string
}

type application struct {
	conf          *config
	authenticator oidcAuth.Authenticator
	session       *scs.SessionManager
	router        chi.Router
	dishRepo      domain.DishRepo
}

func parseConfig() (*config, error) {

	setEnvErr := func(envVarName string) error {
		return fmt.Errorf("missing env var %v", envVarName)
	}

	cfg := config{}

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

	if oidcSecret := os.Getenv(envOIDCSecret); oidcSecret == "" {
		return nil, setEnvErr(envOIDCSecret)
	} else {
		cfg.oidcSecret = oidcSecret
	}

	if oidcCallbackURL := os.Getenv(envOIDCCallbackURL); oidcCallbackURL == "" {
		return nil, setEnvErr(envOIDCCallbackURL)
	} else {
		cfg.oidcCallbackURL = oidcCallbackURL
	}

	if oidcProviderURL := os.Getenv(envOIDCProviderURL); oidcProviderURL == "" {
		return nil, setEnvErr(envOIDCProviderURL)
	} else {
		cfg.oidcProviderURL = oidcProviderURL
	}

	if oidcID := os.Getenv(envOIDCID); oidcID == "" {
		return nil, setEnvErr(envOIDCID)
	} else {
		cfg.oidcID = oidcID
	}

	if botApiToken := os.Getenv(envBotAPIToken); botApiToken == "" {
		return nil, setEnvErr(envBotAPIToken)
	} else {
		cfg.botAPIToken = botApiToken
	}

	cfg.listen = ":80"

	return &cfg, nil
}

func newApplication(cfg *config) (*application, error) {

	//Build Session Storage

	log.Printf("Building session storage...")
	session := scs.New()
	session.Lifetime = 1 * time.Hour
	session.Cookie.Secure = true

	//Build oidc authenticator

	log.Printf("Building auth backend...")
	authStorageAdapter := NewAuthSessionStorageManager(session)
	authenticator, err := oidcAuth.NewDefaultAuthenticator(cfg.oidcProviderURL, cfg.oidcID, cfg.oidcSecret,
		cfg.oidcCallbackURL, "/whoami", "/", authStorageAdapter)
	if err != nil {
		return nil, fmt.Errorf("oidcAuth.NewDefaultAuthenticator : %v", err)
	}

	//Connect to db
	db, err := connectToDB(context.Background(), cfg.dbUser, cfg.dbPW, cfg.dbURL, cfg.dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect do db : %v")
	}

	//Build dish repo
	repo, err := dishRepo.NewMysqlRepo(db)
	if err != nil {
		return nil, fmt.Errorf("failed to build mysql dish repo : %v", err)
	}

	app := application{
		conf:          cfg,
		authenticator: authenticator,
		session:       session,
		dishRepo:      repo,
	}

	app.router, err = app.setupRouter()
	if err != nil {
		return nil, fmt.Errorf("app.setupRouter : %v", err)
	}

	return &app, nil

}

// connectToDB connects to the db waiting some tiem for the db to come online before giving up
func connectToDB(ctx context.Context, dbUser, dbPW, dbURL, dbName string) (*sql.DB, error) {

	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", dbUser, dbPW, dbURL, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open dsn %v : %v", dsn, err)
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

func (app *application) setupRouter() (chi.Router, error) {
	log.Printf("Configuring router...")
	router := chi.NewRouter()

	router.Handle("/callback", http.HandlerFunc(app.authenticator.CallbackHandler))
	router.Handle("/login", http.HandlerFunc(app.authenticator.LoginHandler))
	router.Handle("/logout", http.HandlerFunc(app.authenticator.LogoutHandler))

	//Builder User API for dishes

	userAPiRouter := chi.NewRouter()
	userAPiRouter.Use(app.session.LoadAndSave)
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

	userAPIServer := userAPI.NewHttpServer()
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

	botAPIServer := botAPI.NewService(app.dishRepo)
	botAPIHandlers := botAPI.NewStrictHandler(botAPIServer, nil)
	botAPI.HandlerFromMux(botAPIHandlers, botAPIRouter)
	router.Mount("/botAPI/v1", botAPIRouter)

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
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Printf("Http server performs gracefull shutdown")
			} else {
				log.Printf("Error in HTTP server : %v", err)
			}
		}
	}()

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

	log.Printf("Parsing config...")
	cfg, err := parseConfig()
	if err != nil {
		log.Printf("parseConfig : %v", err)
		return
	}

	log.Printf("Building application...")
	app, err := newApplication(cfg)
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
