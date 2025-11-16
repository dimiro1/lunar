package api

import (
	"encoding/json"
	"net/http"
)

type LoginRequest struct {
	APIKey string `json:"apiKey"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HandleLogin validates the API key and sets an HttpOnly cookie
func HandleLogin(apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Invalid request body",
			})
			return
		}

		// Validate API key using constant-time comparison
		if !isValidAPIKey(req.APIKey, apiKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Invalid API key",
			})
			return
		}

		// Set HttpOnly cookie with 1-day expiration
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    req.APIKey,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400, // 1 day in seconds
		}

		// Set Secure flag if using HTTPS
		if r.TLS != nil {
			cookie.Secure = true
		}

		http.SetCookie(w, cookie)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
		})
	}
}

// HandleLogout clears the authentication cookie
func HandleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Clear the cookie by setting MaxAge to -1
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		}

		http.SetCookie(w, cookie)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
		})
	}
}
