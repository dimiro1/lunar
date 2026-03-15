package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/dimiro1/lunar/internal/token"
)

func TestDeviceAuthStore_CreateAndGet(t *testing.T) {
	ds := NewDeviceAuthStore()

	auth := ds.Create()
	if auth.DeviceCode == "" {
		t.Error("expected DeviceCode to be set")
	}
	if auth.UserCode == "" {
		t.Error("expected UserCode to be set")
	}
	if len(auth.UserCode) != userCodeLength {
		t.Errorf("expected UserCode length %d, got %d", userCodeLength, len(auth.UserCode))
	}
	if auth.Status != PendingAuthStatusPending {
		t.Errorf("expected status pending, got %s", auth.Status)
	}

	// Should be retrievable
	found := ds.Get(auth.DeviceCode)
	if found == nil {
		t.Fatal("expected to find pending auth")
	}
	if found.DeviceCode != auth.DeviceCode {
		t.Errorf("expected DeviceCode %s, got %s", auth.DeviceCode, found.DeviceCode)
	}
}

func TestDeviceAuthStore_GetNotFound(t *testing.T) {
	ds := NewDeviceAuthStore()

	found := ds.Get("nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent device code")
	}
}

func TestDeviceAuthStore_ApproveAndDeny(t *testing.T) {
	ds := NewDeviceAuthStore()

	// Test approve
	auth := ds.Create()
	ok := ds.Approve(auth.DeviceCode, "test-token-value")
	if !ok {
		t.Error("expected Approve to succeed")
	}

	found := ds.Get(auth.DeviceCode)
	if found.Status != PendingAuthStatusApproved {
		t.Errorf("expected status approved, got %s", found.Status)
	}
	if found.Token != "test-token-value" {
		t.Errorf("expected token test-token-value, got %s", found.Token)
	}

	// Test deny
	auth2 := ds.Create()
	ok = ds.Deny(auth2.DeviceCode)
	if !ok {
		t.Error("expected Deny to succeed")
	}

	found2 := ds.Get(auth2.DeviceCode)
	if found2.Status != PendingAuthStatusDenied {
		t.Errorf("expected status denied, got %s", found2.Status)
	}
}

func TestDeviceAuthStore_DoubleApprove(t *testing.T) {
	ds := NewDeviceAuthStore()

	auth := ds.Create()
	ds.Approve(auth.DeviceCode, "token1")

	// Second approve should fail
	ok := ds.Approve(auth.DeviceCode, "token2")
	if ok {
		t.Error("expected second Approve to fail")
	}
}

func TestHandleDeviceRequest(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp DeviceRequestResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.DeviceCode == "" {
		t.Error("expected device_code to be set")
	}
	if resp.UserCode == "" {
		t.Error("expected user_code to be set")
	}
	if resp.ApprovalURL == "" {
		t.Error("expected approval_url to be set")
	}
	if resp.ExpiresIn != 300 {
		t.Errorf("expected expires_in 300, got %d", resp.ExpiresIn)
	}
	if resp.Interval != 5 {
		t.Errorf("expected interval 5, got %d", resp.Interval)
	}
}

func TestHandleDeviceToken_Pending(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// First create a device request
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	wCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(wCreate, reqCreate)

	var createResp DeviceRequestResponse
	if err := json.NewDecoder(wCreate.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Poll for token - should be pending
	req := httptest.NewRequest(http.MethodGet, "/api/auth/device-token?code="+createResp.DeviceCode, nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var tokenResp DeviceTokenResponse
	if err := json.NewDecoder(w.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	if tokenResp.Status != "pending" {
		t.Errorf("expected status pending, got %s", tokenResp.Status)
	}
	if tokenResp.Token != "" {
		t.Error("expected token to be empty for pending status")
	}
}

func TestHandleDeviceToken_NotFound(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/api/auth/device-token?code=nonexistent", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleDeviceToken_MissingCode(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/api/auth/device-token", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleDeviceApproveFlow(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Step 1: Create device request
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	wCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(wCreate, reqCreate)

	var createResp DeviceRequestResponse
	if err := json.NewDecoder(wCreate.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Step 2: Get approve status (requires auth)
	reqStatus := makeAuthRequest(http.MethodGet, "/api/auth/device-approve?code="+createResp.DeviceCode, nil)
	wStatus := httptest.NewRecorder()
	server.Handler().ServeHTTP(wStatus, reqStatus)

	if wStatus.Code != http.StatusOK {
		t.Fatalf("expected status 200 for approve status, got %d: %s", wStatus.Code, wStatus.Body.String())
	}

	var statusResp DeviceApproveStatusResponse
	if err := json.NewDecoder(wStatus.Body).Decode(&statusResp); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}

	if statusResp.Status != "pending" {
		t.Errorf("expected status pending, got %s", statusResp.Status)
	}
	if statusResp.UserCode == "" {
		t.Error("expected user_code to be set")
	}
	if statusResp.DeviceCode != createResp.DeviceCode {
		t.Errorf("expected device_code %s, got %s", createResp.DeviceCode, statusResp.DeviceCode)
	}

	// Step 3: Approve the request (requires auth)
	approveBody, _ := json.Marshal(DeviceApproveRequest{
		DeviceCode: createResp.DeviceCode,
		Action:     "allow",
	})
	reqApprove := makeAuthRequest(http.MethodPost, "/api/auth/device-approve", approveBody)
	wApprove := httptest.NewRecorder()
	server.Handler().ServeHTTP(wApprove, reqApprove)

	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected status 200 for approve, got %d: %s", wApprove.Code, wApprove.Body.String())
	}

	// Step 4: Poll for token - should now be approved
	reqToken := httptest.NewRequest(http.MethodGet, "/api/auth/device-token?code="+createResp.DeviceCode, nil)
	wToken := httptest.NewRecorder()
	server.Handler().ServeHTTP(wToken, reqToken)

	if wToken.Code != http.StatusOK {
		t.Fatalf("expected status 200 for token, got %d: %s", wToken.Code, wToken.Body.String())
	}

	var tokenResp DeviceTokenResponse
	if err := json.NewDecoder(wToken.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	if tokenResp.Status != "approved" {
		t.Errorf("expected status approved, got %s", tokenResp.Status)
	}
	if tokenResp.Token == "" {
		t.Error("expected token to be set")
	}

	// Step 5: Verify the token works for authentication
	reqAuth := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	reqAuth.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	wAuth := httptest.NewRecorder()
	server.Handler().ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusOK {
		t.Errorf("expected status 200 with API token auth, got %d: %s", wAuth.Code, wAuth.Body.String())
	}
}

func TestHandleDeviceDenyFlow(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create device request
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	wCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(wCreate, reqCreate)

	var createResp DeviceRequestResponse
	if err := json.NewDecoder(wCreate.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Deny the request
	denyBody, _ := json.Marshal(DeviceApproveRequest{
		DeviceCode: createResp.DeviceCode,
		Action:     "deny",
	})
	reqDeny := makeAuthRequest(http.MethodPost, "/api/auth/device-approve", denyBody)
	wDeny := httptest.NewRecorder()
	server.Handler().ServeHTTP(wDeny, reqDeny)

	if wDeny.Code != http.StatusOK {
		t.Fatalf("expected status 200 for deny, got %d: %s", wDeny.Code, wDeny.Body.String())
	}

	// Poll for token - should be denied
	reqToken := httptest.NewRequest(http.MethodGet, "/api/auth/device-token?code="+createResp.DeviceCode, nil)
	wToken := httptest.NewRecorder()
	server.Handler().ServeHTTP(wToken, reqToken)

	var tokenResp DeviceTokenResponse
	if err := json.NewDecoder(wToken.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	if tokenResp.Status != "denied" {
		t.Errorf("expected status denied, got %s", tokenResp.Status)
	}
	if tokenResp.Token != "" {
		t.Error("expected token to be empty for denied status")
	}
}

func TestHandleDeviceApprove_RequiresAuth(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	// GET approve status without auth
	req := httptest.NewRequest(http.MethodGet, "/api/auth/device-approve?code=test", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for unauthenticated approve status, got %d", w.Code)
	}

	// POST approve without auth
	body, _ := json.Marshal(DeviceApproveRequest{DeviceCode: "test", Action: "allow"})
	reqPost := httptest.NewRequest(http.MethodPost, "/api/auth/device-approve", bytes.NewReader(body))
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	server.Handler().ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for unauthenticated approve, got %d", wPost.Code)
	}
}

func TestHandleDeviceApprove_InvalidAction(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create device request
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	wCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(wCreate, reqCreate)

	var createResp DeviceRequestResponse
	if err := json.NewDecoder(wCreate.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Try invalid action
	body, _ := json.Marshal(DeviceApproveRequest{
		DeviceCode: createResp.DeviceCode,
		Action:     "invalid",
	})
	req := makeAuthRequest(http.MethodPost, "/api/auth/device-approve", body)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid action, got %d", w.Code)
	}
}

func TestHandleListAPITokens(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// List should be empty initially
	req := makeAuthRequest(http.MethodGet, "/api/tokens", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string][]store.APIToken
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp["tokens"]) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(resp["tokens"]))
	}
}

func TestHandleRevokeAPIToken(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a token via the device flow
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/auth/device-request", nil)
	wCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(wCreate, reqCreate)

	var createResp DeviceRequestResponse
	if err := json.NewDecoder(wCreate.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Approve
	approveBody, _ := json.Marshal(DeviceApproveRequest{
		DeviceCode: createResp.DeviceCode,
		Action:     "allow",
	})
	reqApprove := makeAuthRequest(http.MethodPost, "/api/auth/device-approve", approveBody)
	wApprove := httptest.NewRecorder()
	server.Handler().ServeHTTP(wApprove, reqApprove)

	// Get the token from poll
	reqToken := httptest.NewRequest(http.MethodGet, "/api/auth/device-token?code="+createResp.DeviceCode, nil)
	wToken := httptest.NewRecorder()
	server.Handler().ServeHTTP(wToken, reqToken)

	var tokenResp DeviceTokenResponse
	if err := json.NewDecoder(wToken.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	// List tokens to get the ID
	reqList := makeAuthRequest(http.MethodGet, "/api/tokens", nil)
	wList := httptest.NewRecorder()
	server.Handler().ServeHTTP(wList, reqList)

	var listResp map[string][]store.APIToken
	if err := json.NewDecoder(wList.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(listResp["tokens"]) != 1 {
		t.Fatalf("expected 1 token, got %d", len(listResp["tokens"]))
	}

	tokenID := listResp["tokens"][0].ID

	// Revoke the token
	reqRevoke := makeAuthRequest(http.MethodPost, "/api/tokens/"+tokenID+"/revoke", nil)
	wRevoke := httptest.NewRecorder()
	server.Handler().ServeHTTP(wRevoke, reqRevoke)

	if wRevoke.Code != http.StatusOK {
		t.Fatalf("expected status 200 for revoke, got %d: %s", wRevoke.Code, wRevoke.Body.String())
	}

	// Verify the token no longer works for auth
	reqAuth := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	reqAuth.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	wAuth := httptest.NewRecorder()
	server.Handler().ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 after revocation, got %d", wAuth.Code)
	}
}

func TestHandleRevokeAPIToken_NotFound(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := makeAuthRequest(http.MethodPost, "/api/tokens/nonexistent/revoke", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestAuthMiddleware_WithAPIToken(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create an API token directly in the store
	rawToken := "test-cli-token-12345678901234567890"
	tokenHash := token.Hash(rawToken)

	_, err := database.CreateAPIToken(context.Background(), store.APIToken{
		ID:        "direct_token",
		TokenHash: tokenHash,
		Name:      "Test CLI",
	})
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	// Use the raw token for auth
	req := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 with API token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 with invalid token, got %d", w.Code)
	}
}

func TestAuthMiddleware_AdminKeyStillWorks(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	// Admin API key should still work
	req := makeAuthRequest(http.MethodGet, "/api/functions", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 with admin key, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_CookieStillWorks(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "test-api-key",
	})
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 with cookie auth, got %d: %s", w.Code, w.Body.String())
	}
}
