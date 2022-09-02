package oidcAuth

import (
	"log"
	"net/http"
)

type MockAuthenticator struct {
	session SessionStorage

	urlAfterLogin  string
	urlAfterLogout string

	injectUserOnEachRequest bool
	profile                 UserProfile
}

// NewMockAuthenticator returns and authenticator that simulates the login always loggin in a fixed user
func NewMockAuthenticator(injectUserOnEachRequest bool, urlAfterLogin, urlAfterLogout string,
	storage SessionStorage) *MockAuthenticator {
	return &MockAuthenticator{
		session:                 storage,
		urlAfterLogin:           urlAfterLogin,
		urlAfterLogout:          urlAfterLogout,
		injectUserOnEachRequest: injectUserOnEachRequest,
		profile:                 UserProfile{Email: "testUser@some.domain"},
	}
}

func (m MockAuthenticator) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Mock CheckSession called")
		if m.injectUserOnEachRequest {
			log.Printf("CheckSession injecting user")
			if err := m.session.StoreProfile(r.Context(), SessionKeyProfile, m.profile); err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				log.Printf("failed to store %v to session : %v", SessionKeyProfile, err)
				return
			}
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

	http.Redirect(w, r, m.urlAfterLogin, http.StatusSeeOther)
}

func (m MockAuthenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock callback handler jumping to CallBackHandler")
	m.CallbackHandler(w, r)

}

func (m MockAuthenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := m.session.Destroy(r.Context()); err != nil {
		log.Printf("Failed to destory session: %v", err)
	} else {
		log.Printf("LogoutHandler: cleared session")
	}
	http.Redirect(w, r, m.urlAfterLogout, http.StatusTemporaryRedirect)
}
