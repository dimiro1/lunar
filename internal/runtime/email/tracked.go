package email

import (
	"encoding/json"
	"time"

	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/store"
)

// TrackedSendResult contains both the response and tracking information.
type TrackedSendResult struct {
	Response  *email.SendResponse
	TrackData email.TrackRequest
	Error     error
}

// TrackedClient wraps an email.Client and automatically tracks all requests.
// This implements the Decorator pattern.
type TrackedClient struct {
	client      email.Client
	tracker     email.Tracker
	executionID string
}

// NewTrackedClient creates a TrackedClient that wraps the given client and tracks requests.
func NewTrackedClient(client email.Client, tracker email.Tracker, executionID string) *TrackedClient {
	return &TrackedClient{
		client:      client,
		tracker:     tracker,
		executionID: executionID,
	}
}

// Send sends an email with automatic tracking.
// It measures duration, captures request/response data, and tracks the result.
func (tc *TrackedClient) Send(functionID string, req email.SendRequest) (*email.SendResponse, error) {
	result := tc.SendWithTracking(functionID, req)
	return result.Response, result.Error
}

// SendWithTracking sends an email and returns both the response and tracking data.
// This is useful when you need access to the tracking information.
func (tc *TrackedClient) SendWithTracking(functionID string, req email.SendRequest) TrackedSendResult {
	startTime := time.Now()
	resp, err := tc.client.Send(functionID, req)
	durationMs := time.Since(startTime).Milliseconds()

	// Get request JSON for tracking
	var requestJSON string
	if resp != nil {
		requestJSON = resp.RequestJSON
	}

	trackReq := email.TrackRequest{
		From:        req.From,
		To:          req.To,
		Subject:     req.Subject,
		HasText:     req.Text != "",
		HasHTML:     req.HTML != "",
		RequestJSON: requestJSON,
		DurationMs:  durationMs,
	}

	if err != nil {
		errMsg := err.Error()
		trackReq.Status = store.EmailRequestStatusError
		trackReq.ErrorMessage = &errMsg

		if tc.tracker != nil {
			tc.tracker.Track(tc.executionID, trackReq)
		}

		return TrackedSendResult{
			Response:  nil,
			TrackData: trackReq,
			Error:     err,
		}
	}

	// Build response JSON
	responseJSON, _ := json.Marshal(map[string]string{"id": resp.ID})
	responseJSONStr := string(responseJSON)

	trackReq.Status = store.EmailRequestStatusSuccess
	trackReq.EmailID = &resp.ID
	trackReq.ResponseJSON = &responseJSONStr

	if tc.tracker != nil {
		tc.tracker.Track(tc.executionID, trackReq)
	}

	return TrackedSendResult{
		Response:  resp,
		TrackData: trackReq,
		Error:     nil,
	}
}
