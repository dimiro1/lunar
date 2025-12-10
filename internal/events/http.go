package events

// HTTPEvent represents an incoming HTTP request
type HTTPEvent struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	RelativePath string            `json:"relativePath"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	Query        map[string]string `json:"query"`
}

// Type returns the event type for HTTPEvent
func (h HTTPEvent) Type() EventType {
	return EventTypeHTTP
}

// HTTPResponse represents the HTTP response from a Lua function handler
type HTTPResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}
