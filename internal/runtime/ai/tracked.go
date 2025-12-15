package ai

import (
	"time"

	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/store"
)

// TrackedChatResult contains both the response and tracking information.
type TrackedChatResult struct {
	Response  *ai.ChatResponse
	TrackData ai.TrackRequest
	Error     error
}

// TrackedClient wraps an ai.Client and automatically tracks all requests.
// This implements the Decorator pattern.
type TrackedClient struct {
	client      ai.Client
	tracker     ai.Tracker
	executionID string
}

// NewTrackedClient creates a TrackedClient that wraps the given client and tracks requests.
func NewTrackedClient(client ai.Client, tracker ai.Tracker, executionID string) *TrackedClient {
	return &TrackedClient{
		client:      client,
		tracker:     tracker,
		executionID: executionID,
	}
}

// Chat executes a chat request with automatic tracking.
// It measures duration, captures request/response data, and tracks the result.
func (tc *TrackedClient) Chat(functionID string, req ai.ChatRequest) (*ai.ChatResponse, error) {
	result := tc.ChatWithTracking(functionID, req)
	return result.Response, result.Error
}

// ChatWithTracking executes a chat request and returns both the response and tracking data.
// This is useful when you need access to the tracking information.
func (tc *TrackedClient) ChatWithTracking(functionID string, req ai.ChatRequest) TrackedChatResult {
	trackReq := ai.TrackRequest{
		Provider: req.Provider,
		Model:    req.Model,
	}

	startTime := time.Now()
	response, err := tc.client.Chat(functionID, req)
	trackReq.DurationMs = time.Since(startTime).Milliseconds()

	// Capture tracking info from response even on error (if available)
	if response != nil {
		trackReq.Endpoint = response.Endpoint
		trackReq.RequestJSON = response.RequestJSON
		if response.ResponseJSON != "" {
			trackReq.ResponseJSON = &response.ResponseJSON
		}
	}

	if err != nil {
		errMsg := err.Error()
		trackReq.Status = store.AIRequestStatusError
		trackReq.ErrorMessage = &errMsg

		// Track the error
		if tc.tracker != nil {
			tc.tracker.Track(tc.executionID, trackReq)
		}

		return TrackedChatResult{
			Response:  nil,
			TrackData: trackReq,
			Error:     err,
		}
	}

	trackReq.Status = store.AIRequestStatusSuccess
	trackReq.InputTokens = &response.Usage.InputTokens
	trackReq.OutputTokens = &response.Usage.OutputTokens

	// Track success
	if tc.tracker != nil {
		tc.tracker.Track(tc.executionID, trackReq)
	}

	return TrackedChatResult{
		Response:  response,
		TrackData: trackReq,
		Error:     nil,
	}
}
