package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
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
	envOIDCSecret      = "OIDC_SECRET"
	envOIDCCallbackURL = "OIDC_CALLBACK_URL"
	envOIDCProviderURL = "OIDC_PROVIDER_URL"
	envOIDCID          = "OIDC_ID"
)

type config struct {

	//OIDC Config

	oidcSecret      string
	oidcCallbackURL string
	oidcProviderURL string
	oidcID          string

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
}

func parseConfig() (*config, error) {

	setEnvErr := func(envVarName string) error {
		return fmt.Errorf("missing env var %v", envVarName)
	}

	cfg := config{}

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

	app := application{
		conf:          cfg,
		authenticator: authenticator,
		session:       session,
	}

	app.router, err = app.setupRouter()
	if err != nil {
		return nil, fmt.Errorf("app.setupRouter : %v", err)
	}

	return &app, nil

}

func (app *application) setupRouter() (chi.Router, error) {
	log.Printf("Configuring router...")
	router := chi.NewRouter()
	router.Use(app.session.LoadAndSave)
	router.Use(app.authenticator.CheckSession)

	router.Handle("/callback", http.HandlerFunc(app.authenticator.CallbackHandler))
	router.Handle("/login", http.HandlerFunc(app.authenticator.LoginHandler))
	router.Handle("/logout", http.HandlerFunc(app.authenticator.LogoutHandler))

	//Builder User API for dishes
	swagger, err := userAPI.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec : %v", err)
	}
	swagger.Servers = nil

	userAPiRouter := chi.NewRouter()
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
