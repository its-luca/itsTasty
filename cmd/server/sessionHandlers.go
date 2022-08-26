package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

// processAuthentication, checks if session contains access token and tries refresh if it's expired.
// If token exists contextKeyIsAuthenticated is placed in context. DOES NOT ABORT ON NO SESSION
func (app *application) processAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("processAuthentication middleware was called")
		//check if session contains token
		accessToken := app.session.GetString(r.Context(), "access_token")
		if accessToken == "" {
			log.Printf("invalid sessions")
			ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		//try to refresh token if expired
		expiry := app.session.GetTime(r.Context(), "expiry")
		if expiry.Before(time.Now()) {
			log.Printf("Access token expired, trying to refresh")
			token, err := app.authenticator.Refresh(r.Context(), app.session.GetString(r.Context(), "refresh_token"))
			if err != nil {
				log.Printf("refresh failed: %v", err)
				if err := app.session.Clear(r.Context()); err != nil {
					fmt.Printf("failed to clear session: %v", err)
				} else {
					fmt.Printf("cleared session")
				}
				ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, false)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
			//successful refresh, , update session
			app.session.Put(r.Context(), "access_token", token.AccessToken)
			app.session.Put(r.Context(), "expiry", token.Expiry)
			log.Printf("refreshed session, expires in %v", token.Expiry)
		}

		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// callbackHandler, handles the callback from the OpendID Connect workflow
func (app *application) callbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("callbackHandler was called")

	if r.URL.Query().Get("state") != app.session.GetString(r.Context(), "state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	log.Printf("state check passed, doing exchange...")

	token, err := app.authenticator.Exchange(context.TODO(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("no token found: %v", err)
		http.Error(w, "Problems communicating with login server", http.StatusInternalServerError)
		return
	}

	log.Printf("Done\n extracting token...")

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	log.Printf("Done!\nVerifiyng token...")

	idToken, err := app.authenticator.Verify(context.TODO(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Done!\nParsing User Info...")

	// Getting now the userInfo

	//TODO: type profile
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Done. Profile is %+v\n", profile)

	log.Printf("token type: %v. Expiry: %v\n", token.TokenType, token.Expiry)

	app.session.Put(r.Context(), "access_token", token.AccessToken)
	app.session.Put(r.Context(), "expiry", token.Expiry)
	app.session.Put(r.Context(), "refresh_token", token.RefreshToken)
	app.session.Put(r.Context(), "profile", profile)

	http.Redirect(w, r, "/whoami", http.StatusSeeOther)
}

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler was called")
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	app.session.Put(r.Context(), "state", state)

	http.Redirect(w, r, app.authenticator.GetLoginURL(state), http.StatusTemporaryRedirect)
}

func (app *application) logoutHandler(_ http.ResponseWriter, r *http.Request) {

	if err := app.session.Clear(r.Context()); err != nil {
		log.Printf("Failed to clear local session: %v", err)
	}
	log.Printf("logoutHandler: cleared session")
	//TODO: redirect somewhere useful
}
