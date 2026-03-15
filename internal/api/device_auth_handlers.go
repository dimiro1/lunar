package api

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/dimiro1/lunar/internal/token"
	"github.com/rs/xid"
)

const (
	deviceCodeExpiry = 5 * time.Minute
	userCodeChars    = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no I, O, 0, 1 to avoid confusion
	userCodeLength   = 8
)

// PendingAuthStatus represents the status of a pending device auth request
type PendingAuthStatus string

const (
	PendingAuthStatusPending  PendingAuthStatus = "pending"
	PendingAuthStatusApproved PendingAuthStatus = "approved"
	PendingAuthStatusDenied   PendingAuthStatus = "denied"
)

// PendingAuth represents a pending device authorization request
type PendingAuth struct {
	DeviceCode string
	UserCode   string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Status     PendingAuthStatus
	Token      string // raw token, set on approval
}

// DeviceAuthStore is a thread-safe in-memory store for pending device auth requests.
// Pending auths are stored in memory rather than the database because they are
// short-lived (5 minute TTL) and do not need to survive server restarts.
type DeviceAuthStore struct {
	mu      sync.Mutex
	pending map[string]*PendingAuth // keyed by device_code
}

// NewDeviceAuthStore creates a new DeviceAuthStore
func NewDeviceAuthStore() *DeviceAuthStore {
	return &DeviceAuthStore{
		pending: make(map[string]*PendingAuth),
	}
}

// cleanupExpired removes all expired entries from the pending map.
// Must be called with s.mu held.
func (s *DeviceAuthStore) cleanupExpired() {
	now := time.Now()
	for code, auth := range s.pending {
		if now.After(auth.ExpiresAt) {
			delete(s.pending, code)
		}
	}
}

// Create generates a new pending auth request and returns it
func (s *DeviceAuthStore) Create() *PendingAuth {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpired()

	now := time.Now()
	auth := &PendingAuth{
		DeviceCode: xid.New().String(),
		UserCode:   generateUserCode(),
		CreatedAt:  now,
		ExpiresAt:  now.Add(deviceCodeExpiry),
		Status:     PendingAuthStatusPending,
	}

	s.pending[auth.DeviceCode] = auth
	return auth
}

// Get returns a pending auth by device code, or nil if not found/expired
func (s *DeviceAuthStore) Get(deviceCode string) *PendingAuth {
	s.mu.Lock()
	defer s.mu.Unlock()

	auth, ok := s.pending[deviceCode]
	if !ok {
		return nil
	}

	if time.Now().After(auth.ExpiresAt) {
		delete(s.pending, deviceCode)
		return nil
	}

	return auth
}

// getPending returns a pending auth by device code if it exists, is not expired,
// and is still in pending status. Must be called with s.mu held.
func (s *DeviceAuthStore) getPending(deviceCode string) *PendingAuth {
	auth, ok := s.pending[deviceCode]
	if !ok || time.Now().After(auth.ExpiresAt) || auth.Status != PendingAuthStatusPending {
		return nil
	}
	return auth
}

// Approve marks a pending auth as approved with the given token
func (s *DeviceAuthStore) Approve(deviceCode string, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	auth := s.getPending(deviceCode)
	if auth == nil {
		return false
	}

	auth.Status = PendingAuthStatusApproved
	auth.Token = token
	return true
}

// Deny marks a pending auth as denied
func (s *DeviceAuthStore) Deny(deviceCode string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	auth := s.getPending(deviceCode)
	if auth == nil {
		return false
	}

	auth.Status = PendingAuthStatusDenied
	return true
}

func generateUserCode() string {
	code := make([]byte, userCodeLength)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(userCodeChars))))
		code[i] = userCodeChars[n.Int64()]
	}
	return string(code)
}

// HandleDeviceRequest handles POST /api/auth/device-request
// No auth required - the CLI is not yet authenticated
func HandleDeviceRequest(deviceStore *DeviceAuthStore, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := deviceStore.Create()

		writeJSON(w, http.StatusOK, DeviceRequestResponse{
			DeviceCode:  auth.DeviceCode,
			UserCode:    auth.UserCode,
			ApprovalURL: baseURL + "/#!/device-approve/" + auth.DeviceCode,
			ExpiresIn:   int(deviceCodeExpiry.Seconds()),
			Interval:    5,
		})
	}
}

// HandleDeviceApproveStatus handles GET /api/auth/device-approve?code=<device_code>
// Auth required - returns pending request info for the SPA to render
func HandleDeviceApproveStatus(deviceStore *DeviceAuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			writeError(w, http.StatusBadRequest, "Missing code parameter")
			return
		}

		auth := deviceStore.Get(code)
		if auth == nil {
			writeError(w, http.StatusNotFound, "Device code not found or expired")
			return
		}

		writeJSON(w, http.StatusOK, DeviceApproveStatusResponse{
			DeviceCode: auth.DeviceCode,
			UserCode:   auth.UserCode,
			Status:     string(auth.Status),
			ExpiresAt:  auth.ExpiresAt.Unix(),
		})
	}
}

// HandleDeviceApprove handles POST /api/auth/device-approve
// Auth required - user approves or denies the request
func HandleDeviceApprove(deviceStore *DeviceAuthStore, db store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeviceApproveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.DeviceCode == "" {
			writeError(w, http.StatusBadRequest, "Missing device_code")
			return
		}

		switch req.Action {
		case "allow":
			rawToken, err := token.Generate()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to generate token")
				return
			}

			tokenHash := token.Hash(rawToken)

			auth := deviceStore.Get(req.DeviceCode)
			if auth == nil {
				writeError(w, http.StatusNotFound, "Device code not found or expired")
				return
			}

			_, err = db.CreateAPIToken(r.Context(), store.APIToken{
				ID:        xid.New().String(),
				TokenHash: tokenHash,
				Name:      "CLI (" + auth.UserCode + ")",
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to create token")
				return
			}

			if !deviceStore.Approve(req.DeviceCode, rawToken) {
				writeError(w, http.StatusNotFound, "Device code not found or expired")
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"success": true})

		case "deny":
			if !deviceStore.Deny(req.DeviceCode) {
				writeError(w, http.StatusNotFound, "Device code not found or expired")
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"success": true})

		default:
			writeError(w, http.StatusBadRequest, "Invalid action, must be 'allow' or 'deny'")
		}
	}
}

// HandleDeviceToken handles GET /api/auth/device-token?code=<device_code>
// No auth required - polled by CLI
func HandleDeviceToken(deviceStore *DeviceAuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			writeError(w, http.StatusBadRequest, "Missing code parameter")
			return
		}

		auth := deviceStore.Get(code)
		if auth == nil {
			writeError(w, http.StatusNotFound, "Device code not found or expired")
			return
		}

		switch auth.Status {
		case PendingAuthStatusApproved:
			writeJSON(w, http.StatusOK, DeviceTokenResponse{
				Status: string(PendingAuthStatusApproved),
				Token:  auth.Token,
			})
		case PendingAuthStatusDenied:
			writeJSON(w, http.StatusOK, DeviceTokenResponse{
				Status: string(PendingAuthStatusDenied),
			})
		default:
			writeJSON(w, http.StatusOK, DeviceTokenResponse{
				Status: string(PendingAuthStatusPending),
			})
		}
	}
}

// HandleListAPITokens handles GET /api/tokens
func HandleListAPITokens(db store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokens, err := db.ListAPITokens(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list tokens")
			return
		}

		if tokens == nil {
			tokens = []store.APIToken{}
		}

		writeJSON(w, http.StatusOK, map[string][]store.APIToken{"tokens": tokens})
	}
}

// HandleRevokeAPIToken handles POST /api/tokens/{id}/revoke
func HandleRevokeAPIToken(db store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "Missing token ID")
			return
		}

		err := db.RevokeAPIToken(r.Context(), id)
		if err != nil {
			if err == store.ErrAPITokenNotFound {
				writeError(w, http.StatusNotFound, "Token not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "Failed to revoke token")
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}
