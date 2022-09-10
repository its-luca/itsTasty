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

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Printf("No id_token field in oauth2 token")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Printf("Done!\nVerifiyng token...")

	idToken, err := da.Verify(context.TODO(), rawIDToken)
	if err != nil {
		log.Printf("Failed to verify ID toekn : %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Printf("Done!\nParsing User Info...")

	// Getting now the userInfo

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("Failed to parse claims struct %+v : %v", claims, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Printf("Done. Profile is %+v\n", claims)
	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("Expected key %v in claims %+v", "email", claims)
		return
	}
	emailVerified, ok := claims["email_verified"].(bool)
	if !ok {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("Expected key %v in claims %+v", "email_verified", claims)
		return
	}

	if !emailVerified {
		http.Error(w, "Please verify your email", http.StatusBadGateway)
		log.Printf("Rejecting %+v because email is not verified", claims)
		return
	}

	profile := UserProfile{Email: email}

	log.Printf("token type: %v. Expiry: %v\n", token.TokenType, token.Expiry)

	if err := da.session.StoreString(r.Context(), sessionKeyAccessToken, token.AccessToken); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", sessionKeyAccessToken, err)
		return
	}

	if err := da.session.StoreTime(r.Context(), sessionKeyExpiry, token.Expiry); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", sessionKeyExpiry, err)
		return
	}

	if err := da.session.StoreString(r.Context(), sessionKeyRefreshToken, token.RefreshToken); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", sessionKeyRefreshToken, err)
		return
	}

	if err := da.session.StoreProfile(r.Context(), SessionKeyProfile, profile); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to store %v to session : %v", SessionKeyProfile, err)
		return
	}

	//fetch redirect target that was requested during login. we verified that it is within our page
	redirectTo, err := da.session.GetString(r.Context(), sessionKeyRedirectTarget)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("failed to retrieve redirect target from session : %v", err)
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
	redirectAfterLogin := da.callbackURL

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
		redirectAfterLogin = *u
	}

	//Store redirect after login location in session, to be able to a access it in the redirect handler
	if err := da.session.StoreString(r.Context(), sessionKeyRedirectTarget, redirectAfterLogin.String()); err != nil {
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
