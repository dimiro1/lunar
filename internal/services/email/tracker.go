package email

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dimiro1/lunar/internal/masking"
	"github.com/dimiro1/lunar/internal/store"
	"github.com/rs/xid"
)

// TrackRequest contains the data needed to track an email request
type TrackRequest struct {
	From         string
	To           []string
	Subject      string
	HasText      bool
	HasHTML      bool
	RequestJSON  string
	ResponseJSON *string
	Status       store.EmailRequestStatus
	ErrorMessage *string
	EmailID      *string
	DurationMs   int64
}

// Tracker is an interface for tracking email requests
// executionID is used to isolate requests for each function execution
type Tracker interface {
	Track(executionID string, req TrackRequest)
	Requests(executionID string) []store.EmailRequest
	RequestsPaginated(executionID string, limit, offset int) ([]store.EmailRequest, int64)
}

// MemoryTracker is an in-memory implementation of Tracker
type MemoryTracker struct {
	mu       sync.RWMutex
	requests []store.EmailRequest
}

// NewMemoryTracker creates a new in-memory tracker
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		requests: make([]store.EmailRequest, 0),
	}
}

// Track records an email request
func (m *MemoryTracker) Track(executionID string, req TrackRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Mask sensitive data in request/response JSON
	maskedRequestJSON := masking.MaskJSONBody(req.RequestJSON)
	var maskedResponseJSON *string
	if req.ResponseJSON != nil {
		masked := masking.MaskJSONBody(*req.ResponseJSON)
		maskedResponseJSON = &masked
	}

	emailReq := store.EmailRequest{
		ID:           xid.New().String(),
		ExecutionID:  executionID,
		From:         req.From,
		To:           req.To,
		Subject:      req.Subject,
		HasText:      req.HasText,
		HasHTML:      req.HasHTML,
		RequestJSON:  maskedRequestJSON,
		ResponseJSON: maskedResponseJSON,
		Status:       req.Status,
		ErrorMessage: req.ErrorMessage,
		EmailID:      req.EmailID,
		DurationMs:   req.DurationMs,
		CreatedAt:    time.Now().Unix(),
	}

	m.requests = append(m.requests, emailReq)
}

// Requests returns all email requests for the specified executionID
func (m *MemoryTracker) Requests(executionID string) []store.EmailRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make([]store.EmailRequest, 0)
	for _, req := range m.requests {
		if req.ExecutionID == executionID {
			requests = append(requests, req)
		}
	}
	return requests
}

// RequestsPaginated returns paginated email requests for the specified executionID
func (m *MemoryTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.EmailRequest, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter requests by executionID
	filtered := make([]store.EmailRequest, 0)
	for _, req := range m.requests {
		if req.ExecutionID == executionID {
			filtered = append(filtered, req)
		}
	}

	total := int64(len(filtered))

	// Apply pagination
	if offset >= len(filtered) {
		return []store.EmailRequest{}, total
	}

	end := min(offset+limit, len(filtered))

	return filtered[offset:end], total
}

// Clear removes all tracked requests
func (m *MemoryTracker) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = make([]store.EmailRequest, 0)
}

// SQLiteTracker is a SQLite-backed implementation of Tracker
type SQLiteTracker struct {
	db *sql.DB
}

// NewSQLiteTracker creates a new SQLite-backed tracker
func NewSQLiteTracker(db *sql.DB) *SQLiteTracker {
	return &SQLiteTracker{db: db}
}

// Track records an email request
func (s *SQLiteTracker) Track(executionID string, req TrackRequest) {
	// Mask sensitive data in request/response JSON
	maskedRequestJSON := masking.MaskJSONBody(req.RequestJSON)
	var maskedResponseJSON *string
	if req.ResponseJSON != nil {
		masked := masking.MaskJSONBody(*req.ResponseJSON)
		maskedResponseJSON = &masked
	}

	id := xid.New().String()
	toAddresses := strings.Join(req.To, ",")

	hasText := 0
	if req.HasText {
		hasText = 1
	}
	hasHTML := 0
	if req.HasHTML {
		hasHTML = 1
	}

	_, err := s.db.Exec(
		`INSERT INTO email_requests
		(id, execution_id, from_address, to_addresses, subject, has_text, has_html,
		 request_json, response_json, status, error_message, email_id, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, executionID, req.From, toAddresses, req.Subject, hasText, hasHTML,
		maskedRequestJSON, maskedResponseJSON, req.Status, req.ErrorMessage,
		req.EmailID, req.DurationMs, time.Now().Unix(),
	)
	if err != nil {
		// Log error but don't fail the execution
		fmt.Printf("Failed to track email request: %v\n", err)
	}
}

// Requests returns all email requests for the specified executionID
func (s *SQLiteTracker) Requests(executionID string) []store.EmailRequest {
	rows, err := s.db.Query(
		`SELECT id, execution_id, from_address, to_addresses, subject, has_text, has_html,
		        request_json, response_json, status, error_message, email_id, duration_ms, created_at
		 FROM email_requests WHERE execution_id = ? ORDER BY created_at`,
		executionID,
	)
	if err != nil {
		return []store.EmailRequest{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanRequests(rows)
}

// RequestsPaginated returns paginated email requests for the specified executionID
func (s *SQLiteTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.EmailRequest, int64) {
	// Get total count
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM email_requests WHERE execution_id = ?", executionID).Scan(&total)
	if err != nil {
		return []store.EmailRequest{}, 0
	}

	// Get paginated requests
	rows, err := s.db.Query(
		`SELECT id, execution_id, from_address, to_addresses, subject, has_text, has_html,
		        request_json, response_json, status, error_message, email_id, duration_ms, created_at
		 FROM email_requests WHERE execution_id = ? ORDER BY created_at LIMIT ? OFFSET ?`,
		executionID, limit, offset,
	)
	if err != nil {
		return []store.EmailRequest{}, total
	}
	defer func() { _ = rows.Close() }()

	return s.scanRequests(rows), total
}

// scanRequests is a helper to scan rows into EmailRequest slice
func (s *SQLiteTracker) scanRequests(rows *sql.Rows) []store.EmailRequest {
	requests := make([]store.EmailRequest, 0)
	for rows.Next() {
		var req store.EmailRequest
		var toAddresses string
		var responseJSON, errorMessage, emailID sql.NullString
		var hasText, hasHTML int

		if err := rows.Scan(
			&req.ID, &req.ExecutionID, &req.From, &toAddresses, &req.Subject,
			&hasText, &hasHTML, &req.RequestJSON, &responseJSON, &req.Status,
			&errorMessage, &emailID, &req.DurationMs, &req.CreatedAt,
		); err != nil {
			continue
		}

		req.To = strings.Split(toAddresses, ",")
		req.HasText = hasText == 1
		req.HasHTML = hasHTML == 1

		if responseJSON.Valid {
			req.ResponseJSON = &responseJSON.String
		}
		if errorMessage.Valid {
			req.ErrorMessage = &errorMessage.String
		}
		if emailID.Valid {
			req.EmailID = &emailID.String
		}

		requests = append(requests, req)
	}
	return requests
}

// EmailParamsToJSON converts the email parameters to a JSON string for logging
func EmailParamsToJSON(from string, to []string, subject, text, html, replyTo string, cc, bcc []string, scheduledAt string, headers map[string]string, tags []map[string]string) string {
	params := map[string]any{
		"from":    from,
		"to":      to,
		"subject": subject,
	}
	if text != "" {
		params["text"] = text
	}
	if html != "" {
		params["html"] = html
	}
	if replyTo != "" {
		params["reply_to"] = replyTo
	}
	if len(cc) > 0 {
		params["cc"] = cc
	}
	if len(bcc) > 0 {
		params["bcc"] = bcc
	}
	if scheduledAt != "" {
		params["scheduled_at"] = scheduledAt
	}
	if len(headers) > 0 {
		params["headers"] = headers
	}
	if len(tags) > 0 {
		params["tags"] = tags
	}

	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}
