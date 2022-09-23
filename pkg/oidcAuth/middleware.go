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
			if err := da.session.Destroy(r.Context()); err != nil {
				fmt.Printf("failed to clear session: %v", err)
				return
			} else {
				fmt.Printf("cleared session")
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

			err := da.Refresh(r.Context())
			if err != nil {
				log.Printf("refresh failed: %v", err)
				if err := da.session.Destroy(r.Context()); err != nil {
					fmt.Printf("failed to clear session: %v", err)
					return
				} else {
					fmt.Printf("cleared session")
				}
				next.ServeHTTP(w, r)
				return
			}

			log.Printf("sucesfully refreshed session")
		}

		log.Printf("Check session middleware: allowed")
		next.ServeHTTP(w, r)
	})
}
