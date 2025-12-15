package ai

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/dimiro1/lunar/internal/masking"
	"github.com/dimiro1/lunar/internal/store"
	"github.com/rs/xid"
)

// TrackRequest contains the data needed to track an AI request
type TrackRequest struct {
	Provider     string
	Model        string
	Endpoint     string
	RequestJSON  string
	ResponseJSON *string
	Status       store.AIRequestStatus
	ErrorMessage *string
	InputTokens  *int
	OutputTokens *int
	DurationMs   int64
}

// Tracker is an interface for tracking AI requests
// executionID is used to isolate requests for each function execution
type Tracker interface {
	Track(executionID string, req TrackRequest)
	Requests(executionID string) []store.AIRequest
	RequestsPaginated(executionID string, limit, offset int) ([]store.AIRequest, int64)
}

// MemoryTracker is an in-memory implementation of Tracker
type MemoryTracker struct {
	mu       sync.RWMutex
	requests []store.AIRequest
}

// NewMemoryTracker creates a new in-memory tracker
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		requests: make([]store.AIRequest, 0),
	}
}

// Track records an AI request
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

	aiReq := store.AIRequest{
		ID:           xid.New().String(),
		ExecutionID:  executionID,
		Provider:     req.Provider,
		Model:        req.Model,
		Endpoint:     req.Endpoint,
		RequestJSON:  maskedRequestJSON,
		ResponseJSON: maskedResponseJSON,
		Status:       req.Status,
		ErrorMessage: req.ErrorMessage,
		InputTokens:  req.InputTokens,
		OutputTokens: req.OutputTokens,
		DurationMs:   req.DurationMs,
		CreatedAt:    time.Now().Unix(),
	}

	m.requests = append(m.requests, aiReq)
}

// Requests returns all AI requests for the specified executionID
func (m *MemoryTracker) Requests(executionID string) []store.AIRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make([]store.AIRequest, 0)
	for _, req := range m.requests {
		if req.ExecutionID == executionID {
			requests = append(requests, req)
		}
	}
	return requests
}

// RequestsPaginated returns paginated AI requests for the specified executionID
func (m *MemoryTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.AIRequest, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter requests by executionID
	filtered := make([]store.AIRequest, 0)
	for _, req := range m.requests {
		if req.ExecutionID == executionID {
			filtered = append(filtered, req)
		}
	}

	total := int64(len(filtered))

	// Apply pagination
	if offset >= len(filtered) {
		return []store.AIRequest{}, total
	}

	end := min(offset+limit, len(filtered))

	return filtered[offset:end], total
}

// Clear removes all tracked requests
func (m *MemoryTracker) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = make([]store.AIRequest, 0)
}

// SQLiteTracker is a SQLite-backed implementation of Tracker
type SQLiteTracker struct {
	db *sql.DB
}

// NewSQLiteTracker creates a new SQLite-backed tracker
func NewSQLiteTracker(db *sql.DB) *SQLiteTracker {
	return &SQLiteTracker{db: db}
}

// Track records an AI request
func (s *SQLiteTracker) Track(executionID string, req TrackRequest) {
	// Mask sensitive data in request/response JSON
	maskedRequestJSON := masking.MaskJSONBody(req.RequestJSON)
	var maskedResponseJSON *string
	if req.ResponseJSON != nil {
		masked := masking.MaskJSONBody(*req.ResponseJSON)
		maskedResponseJSON = &masked
	}

	id := xid.New().String()
	_, err := s.db.Exec(
		`INSERT INTO ai_requests
		(id, execution_id, provider, model, endpoint, request_json, response_json,
		 status, error_message, input_tokens, output_tokens, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, executionID, req.Provider, req.Model, req.Endpoint, maskedRequestJSON,
		maskedResponseJSON, req.Status, req.ErrorMessage, req.InputTokens,
		req.OutputTokens, req.DurationMs, time.Now().Unix(),
	)
	if err != nil {
		// Log error but don't fail the execution
		fmt.Printf("Failed to track AI request: %v\n", err)
	}
}

// Requests returns all AI requests for the specified executionID
func (s *SQLiteTracker) Requests(executionID string) []store.AIRequest {
	rows, err := s.db.Query(
		`SELECT id, execution_id, provider, model, endpoint, request_json, response_json,
		        status, error_message, input_tokens, output_tokens, duration_ms, created_at
		 FROM ai_requests WHERE execution_id = ? ORDER BY created_at`,
		executionID,
	)
	if err != nil {
		return []store.AIRequest{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanRequests(rows)
}

// RequestsPaginated returns paginated AI requests for the specified executionID
func (s *SQLiteTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.AIRequest, int64) {
	// Get total count
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ai_requests WHERE execution_id = ?", executionID).Scan(&total)
	if err != nil {
		return []store.AIRequest{}, 0
	}

	// Get paginated requests
	rows, err := s.db.Query(
		`SELECT id, execution_id, provider, model, endpoint, request_json, response_json,
		        status, error_message, input_tokens, output_tokens, duration_ms, created_at
		 FROM ai_requests WHERE execution_id = ? ORDER BY created_at LIMIT ? OFFSET ?`,
		executionID, limit, offset,
	)
	if err != nil {
		return []store.AIRequest{}, total
	}
	defer func() { _ = rows.Close() }()

	return s.scanRequests(rows), total
}

// scanRequests is a helper to scan rows into AIRequest slice
func (s *SQLiteTracker) scanRequests(rows *sql.Rows) []store.AIRequest {
	requests := make([]store.AIRequest, 0)
	for rows.Next() {
		var req store.AIRequest
		var responseJSON, errorMessage sql.NullString
		var inputTokens, outputTokens sql.NullInt64

		if err := rows.Scan(
			&req.ID, &req.ExecutionID, &req.Provider, &req.Model, &req.Endpoint,
			&req.RequestJSON, &responseJSON, &req.Status, &errorMessage,
			&inputTokens, &outputTokens, &req.DurationMs, &req.CreatedAt,
		); err != nil {
			continue
		}

		if responseJSON.Valid {
			req.ResponseJSON = &responseJSON.String
		}
		if errorMessage.Valid {
			req.ErrorMessage = &errorMessage.String
		}
		if inputTokens.Valid {
			tokens := int(inputTokens.Int64)
			req.InputTokens = &tokens
		}
		if outputTokens.Valid {
			tokens := int(outputTokens.Int64)
			req.OutputTokens = &tokens
		}

		requests = append(requests, req)
	}
	return requests
}
