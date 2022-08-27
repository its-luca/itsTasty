package oidcAuth

import (
	"context"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

type Authenticator interface {
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
	GetLoginURL(state string) string
	Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	CheckSession(next http.Handler) http.Handler
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

const sessionKeyState = "oidcAuthState"
const sessionKeyAccessToken = "oidcAuthAccessToken"
const sessionKeyExpiry = "oidcAuthExpiry"
const sessionKeyRefreshToken = "oidcAuthRefreshToken"
const SessionKeyProfile = "oidcAuthProfile"

type SessionStorage interface {
	StoreString(ctx context.Context, key, value string) error
	GetString(ctx context.Context, key string) (string, error)
	StoreTime(ctx context.Context, key string, value time.Time) error
	GetTime(ctx context.Context, key string) (time.Time, error)
	StoreProfile(ctx context.Context, key string, profile UserProfile) error
	GetProfile(ctx context.Context, key string) (UserProfile, error)
	ClearEntry(ctx context.Context, key string) error
	Destroy(ctx context.Context) error
}

type UserProfile struct {
	Email string `json:"email,omitempty"`
}

type DefaultAuthenticator struct {
	Provider    *oidc.Provider
	Config      oauth2.Config
	Ctx         context.Context
	ProviderURL string
	session     SessionStorage

	urlAfterLogin  string
	urlAfterLogout string
}

func NewDefaultAuthenticator(providerURL, clientID, clientSecret, callbackURL, urlAfterLogin, urlAfterLogout string,
	storage SessionStorage) (*DefaultAuthenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Printf("failed to get provider %v: %v", providerURL, err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &DefaultAuthenticator{
		Provider:       provider,
		Config:         conf,
		Ctx:            ctx,
		ProviderURL:    providerURL,
		session:        storage,
		urlAfterLogin:  urlAfterLogin,
		urlAfterLogout: urlAfterLogout,
	}, nil
}

func (da *DefaultAuthenticator) Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	ts := da.Config.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	return ts.Token()

}

func (da *DefaultAuthenticator) GetLoginURL(state string) string {
	return da.Config.AuthCodeURL(state)

}
func (da *DefaultAuthenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return da.Config.Exchange(ctx, code)
}
func (da *DefaultAuthenticator) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: da.Config.ClientID,
	}
	return da.Provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}
