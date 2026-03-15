package api

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/dimiro1/lunar/internal/token"
)

// AuthMiddleware validates authentication via cookie or Bearer token.
// It checks the admin API key first, then falls back to API tokens in the database.
func AuthMiddleware(apiKey string, db store.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check cookie first (always uses admin API key)
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
					bearerToken := parts[1]

					// First check against admin API key
					if isValidAPIKey(bearerToken, apiKey) {
						next.ServeHTTP(w, r)
						return
					}

					// Then check against API tokens in the database
					tokenHash := token.Hash(bearerToken)
					apiToken, err := db.GetAPITokenByHash(r.Context(), tokenHash)
					if err == nil {
						// Valid token found, update last_used asynchronously
						go func() {
							if err := db.UpdateAPITokenLastUsed(r.Context(), apiToken.ID, time.Now().Unix()); err != nil {
								slog.Error("Failed to update API token last_used", "error", err)
							}
						}()
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
