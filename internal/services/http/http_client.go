package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Headers represents HTTP headers as a map of string key-value pairs
type Headers map[string]string

// Query represents URL query parameters
type Query map[string]string

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers Headers
	Query   Query
	Body    string
}

// Error represents an HTTP error with additional context
type Error struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Client is an interface for making HTTP requests
type Client interface {
	Get(req Request) (Response, error)
	Post(req Request) (Response, error)
	Put(req Request) (Response, error)
	Patch(req Request) (Response, error)
	Delete(req Request) (Response, error)
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    Headers
	Body       string
}

// IsSuccess returns true if the status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError returns true if the status code is 4xx or 5xx
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// DefaultClient is the default implementation of Client
type DefaultClient struct {
	client *http.Client
}

// NewDefaultClient creates a new default HTTP client with timeout and connection pooling
func NewDefaultClient() *DefaultClient {
	return &DefaultClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Get performs an HTTP GET request
func (c *DefaultClient) Get(req Request) (Response, error) {
	return c.doHTTPRequest("GET", req)
}

// Post performs an HTTP POST request
func (c *DefaultClient) Post(req Request) (Response, error) {
	return c.doHTTPRequest("POST", req)
}

// Put performs an HTTP PUT request
func (c *DefaultClient) Put(req Request) (Response, error) {
	return c.doHTTPRequest("PUT", req)
}

// Patch performs an HTTP PATCH request
func (c *DefaultClient) Patch(req Request) (Response, error) {
	return c.doHTTPRequest("PATCH", req)
}

// Delete performs an HTTP DELETE request
func (c *DefaultClient) Delete(req Request) (Response, error) {
	return c.doHTTPRequest("DELETE", req)
}

// doHTTPRequest builds and executes an HTTP request
func (c *DefaultClient) doHTTPRequest(method string, httpReq Request) (Response, error) {
	// Parse and build URL with query parameters
	parsedURL, err := url.Parse(httpReq.URL)
	if err != nil {
		return Response{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	if len(httpReq.Query) > 0 {
		q := parsedURL.Query()
		for key, value := range httpReq.Query {
			q.Add(key, value)
		}
		parsedURL.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	if httpReq.Body != "" {
		bodyReader = strings.NewReader(httpReq.Body)
	}

	req, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers (nil-safe)
	for key, value := range httpReq.Headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

// doRequest executes the HTTP request and converts the response
func (c *DefaultClient) doRequest(req *http.Request) (Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert headers
	headers := make(Headers)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
	}, nil
}

// FakeClient is a stub implementation of Client for testing
type FakeClient struct {
	// Responses maps method+URL to a Response
	Responses map[string]Response
	// Errors maps method+URL to an error
	Errors map[string]error
	// Requests stores all requests made (for verification)
	Requests []Request
}

// NewFakeClient creates a new fake HTTP client for testing
func NewFakeClient() *FakeClient {
	return &FakeClient{
		Responses: make(map[string]Response),
		Errors:    make(map[string]error),
		Requests:  make([]Request, 0),
	}
}

// SetResponse configures a fake response for a given method and URL
func (f *FakeClient) SetResponse(method, url string, response Response) {
	key := method + ":" + url
	f.Responses[key] = response
}

// SetError configures a fake error for a given method and URL
func (f *FakeClient) SetError(method, url string, err error) {
	key := method + ":" + url
	f.Errors[key] = err
}

// Get performs a fake HTTP GET request
func (f *FakeClient) Get(req Request) (Response, error) {
	return f.do("GET", req)
}

// Post performs a fake HTTP POST request
func (f *FakeClient) Post(req Request) (Response, error) {
	return f.do("POST", req)
}

// Put performs a fake HTTP PUT request
func (f *FakeClient) Put(req Request) (Response, error) {
	return f.do("PUT", req)
}

// Patch performs a fake HTTP PATCH request
func (f *FakeClient) Patch(req Request) (Response, error) {
	return f.do("PATCH", req)
}

// Delete performs a fake HTTP DELETE request
func (f *FakeClient) Delete(req Request) (Response, error) {
	return f.do("DELETE", req)
}

// do is the internal method that handles fake request processing
func (f *FakeClient) do(method string, req Request) (Response, error) {
	// Store the request for verification
	f.Requests = append(f.Requests, req)

	key := method + ":" + req.URL

	// Check if there's an error configured for this request
	if err, exists := f.Errors[key]; exists {
		return Response{}, err
	}

	// Check if there's a response configured for this request
	if resp, exists := f.Responses[key]; exists {
		return resp, nil
	}

	// Default response if nothing is configured
	return Response{
		StatusCode: 200,
		Headers:    Headers{},
		Body:       "",
	}, nil
}
