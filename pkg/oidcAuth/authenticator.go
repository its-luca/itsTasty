package oidcAuth

import (
	"context"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"log"
)

type Authenticator interface {
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
	GetLoginURL(state string) string
	Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error)
}

type DefaultAuthenticator struct {
	Provider    *oidc.Provider
	Config      oauth2.Config
	Ctx         context.Context
	ProviderURL string
}

func (d *DefaultAuthenticator) Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	ts := d.Config.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	return ts.Token()

}

func (d *DefaultAuthenticator) GetLoginURL(state string) string {
	return d.Config.AuthCodeURL(state)

}
func (d *DefaultAuthenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return d.Config.Exchange(ctx, code)
}
func (d *DefaultAuthenticator) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: d.Config.ClientID,
	}
	return d.Provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func NewDefaultAuthenticator(providerURL, clientID, clientSecret, callbackURL string) (*DefaultAuthenticator, error) {
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
		Provider:    provider,
		Config:      conf,
		Ctx:         ctx,
		ProviderURL: providerURL,
	}, nil
}
