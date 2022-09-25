package oidcAuth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
)

// CallbackHandler handles the oidc callback, establishing a session, and finally redirects back to the calling application
// The redirect location is expected to be placed into the session key `sessionKeyRedirectTarget` by the login handler
func (da *DefaultAuthenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("CallbackHandler was called")

	stateFromSession, err := da.session.GetString(r.Context(), sessionKeyState)
	if err != nil {
		log.Printf("failed to get key %v from session : %v", sessionKeyState, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("state") != stateFromSession {
		http.Error(w, "", http.StatusBadRequest)
		log.Printf("Invalid state parameter")
		return
	}

	log.Printf("state check passed, doing exchange...")

	token, err := da.Exchange(context.TODO(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("Exchange with login server failed: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Printf("Done\n extracting token...")

	//fetch redirect target that was requested during login. we verified that it is within our page
	redirectTo, err := da.session.GetString(r.Context(), sessionKeyRedirectTarget)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to retrieve redirect target from session : %v", err)
		return
	}

	if err := da.verifyTokenAndStoreInSession(r.Context(), token); err != nil {
		log.Printf("verifyTokenAndStoreInSession failed : %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// LoginHandler initiates the oidc login workflow which will trigger the call to CallbackHandler
func (da *DefaultAuthenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler was called")
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Failed to generate randomness for random state value : %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Store random state in session, so that we can retrieve it in the callback handler to check if it is equal
	//to the provided state
	if err := da.session.StoreString(r.Context(), sessionKeyState, state); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("Failed to save %v in session : %v", sessionKeyState, err)
		return
	}

	// Check if the user requested a specific page to be redirected to after login
	redirectAfterLogin := da.defaultURLAfterLogin

	if target := r.URL.Query().Get("redirectTo"); target != "" {
		u, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Invalid redirectTo URL", http.StatusBadRequest)
			log.Printf("Failed to parse redirectTo URL : %v", err)
			return
		}
		if u.Hostname() != da.callbackURL.Hostname() {
			http.Error(w, "Invalid redirectTo URL", http.StatusBadRequest)
			log.Printf("redirectTo was set to \"%v\" which is outside of our page", u.String())
			return
		}
		redirectAfterLogin = u.String()
	}

	//Store redirect after login location in session, to be able to a access it in the redirect handler
	if err := da.session.StoreString(r.Context(), sessionKeyRedirectTarget, redirectAfterLogin); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("Failed to save value %v for key %v in session : %v", redirectAfterLogin, sessionKeyRedirectTarget, err)
		return
	}

	http.Redirect(w, r, da.GetLoginURL(state), http.StatusTemporaryRedirect)
}

// LogoutHandler destroys the session and redirects to da.urlAfterLogout
func (da *DefaultAuthenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	if err := da.session.Destroy(r.Context()); err != nil {
		log.Printf("Failed to destory session: %v", err)
	} else {
		log.Printf("LogoutHandler: cleared session")
	}
	http.Redirect(w, r, da.urlAfterLogout, http.StatusTemporaryRedirect)
}
