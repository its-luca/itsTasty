package oidcAuth

import (
	"log"
	"net/http"
	"net/url"
)

type MockAuthenticator struct {
	session SessionStorage

	defaultURLAfterLogin string
	urlAfterLogout       string

	profile UserProfile
}

// NewMockAuthenticator returns and authenticator that creates cookie on login, checks existence in CheckSession
// and destroys cookie in logout. Cookie is always for a fixed user
func NewMockAuthenticator(defaultURLAfterLogin, urlAfterLogout string,
	storage SessionStorage) *MockAuthenticator {
	return &MockAuthenticator{
		session:              storage,
		defaultURLAfterLogin: defaultURLAfterLogin,
		urlAfterLogout:       urlAfterLogout,
		profile:              UserProfile{Email: "testUser@some.domain"},
	}
}

func (m MockAuthenticator) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Mock CheckSession called")

		_, err := m.session.GetProfile(r.Context(), SessionKeyProfile)
		if err != nil {
			http.Error(w, "", http.StatusUnauthorized)
			log.Printf("CheckSession: unathorized access")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MockCallbackHandler simply stores a fixed UserProfile under SessionKeyProfile

func (m *MockAuthenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("MockCallbackHandler was called")

	if err := m.session.StoreProfile(r.Context(), SessionKeyProfile, m.profile); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", SessionKeyProfile, err)
		return
	}

	//fetch redirect target that was requested during login
	redirectTo, err := m.session.GetString(r.Context(), sessionKeyRedirectTarget)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to retrieve redirect target from session : %v", err)
		return
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (m *MockAuthenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock LoginHandler called,creating session")
	if err := m.session.StoreProfile(r.Context(), SessionKeyProfile, m.profile); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", SessionKeyProfile, err)
		return
	}

	// Check if the user requested a specific page to be redirected to after login
	redirectAfterLogin := m.defaultURLAfterLogin

	if target := r.URL.Query().Get("redirectTo"); target != "" {
		u, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Invalid redirectTo URL", http.StatusBadRequest)
			log.Printf("Failed to parse redirectTo URL : %v", err)
			return
		}
		redirectAfterLogin = u.String()
	}

	//Store redirect after login location in session, to be able to a access it in the redirect handler
	if err := m.session.StoreString(r.Context(), sessionKeyRedirectTarget, redirectAfterLogin); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("Failed to save value %v for key %v in session : %v", redirectAfterLogin, sessionKeyRedirectTarget, err)
		return
	}

	m.CallbackHandler(w, r)

}

func (m *MockAuthenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := m.session.Destroy(r.Context()); err != nil {
		log.Printf("Failed to destory session: %v", err)
	} else {
		log.Printf("LogoutHandler: cleared session")
	}
	http.Redirect(w, r, m.urlAfterLogout, http.StatusTemporaryRedirect)
}
