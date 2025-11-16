package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// AuthMiddleware validates authentication via cookie or Bearer token
func AuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check cookie first
			if cookie, err := r.Cookie("auth_token"); err == nil {
				if isValidAPIKey(cookie.Value, apiKey) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Expected format: "Bearer {token}"
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					if isValidAPIKey(parts[1], apiKey) {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// No valid authentication found
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
		})
	}
}

// isValidAPIKey uses constant-time comparison to prevent timing attacks
func isValidAPIKey(provided, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
