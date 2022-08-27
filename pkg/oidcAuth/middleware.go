package oidcAuth

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// CheckSession checks if session contains access token and tries refresh if it's expired.
// If token exists UserProfile is guaranteed to be stored as JSON under the key SessionKeyProfile in the session storage
// DOES NOT ABORT ON NO SESSION
func (da *DefaultAuthenticator) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CheckSession middleware was called")
		//check if session contains token
		accessToken, err := da.session.GetString(r.Context(), sessionKeyAccessToken)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("Failed to get %v from session : %v", sessionKeyAccessToken, err)
			return
		}
		if accessToken == "" {
			log.Printf("invalid sessions")
			if err := da.session.ClearEntry(r.Context(), SessionKeyProfile); err != nil {
				log.Printf("failed to clear session SessionKeyProfile from session : %v", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		//try to refresh token if expired
		expiry, err := da.session.GetTime(r.Context(), sessionKeyExpiry)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("Failed to get %v from session : %v", sessionKeyExpiry, err)
			return
		}

		if expiry.Before(time.Now()) {
			log.Printf("Access token expired, trying to refresh")
			refreshToken, err := da.session.GetString(r.Context(), sessionKeyRefreshToken)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				log.Printf("Failed to get %v from session : %v", sessionKeyRefreshToken, err)
				return
			}
			token, err := da.Refresh(r.Context(), refreshToken)
			if err != nil {
				log.Printf("refresh failed: %v", err)
				if err := da.session.Destroy(r.Context()); err != nil {
					fmt.Printf("failed to clear session: %v", err)
				} else {
					fmt.Printf("cleared session")
				}
				next.ServeHTTP(w, r)
				return
			}
			//successful refresh, , update session
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
			log.Printf("refreshed session, expires in %v", token.Expiry)
		}
		next.ServeHTTP(w, r)
	})
}
