package ai

import (
	"encoding/json"
	"fmt"

	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
)

// provider defines the interface for AI provider implementations.
// Each provider (e.g., OpenAI, Anthropic) implements this interface
// to handle provider-specific API details.
type provider interface {
	// defaultEndpoint returns the default API base URL for this provider.
	defaultEndpoint() string
	// urlPath returns the API path to append to the endpoint (e.g., "/v1/chat/completions").
	urlPath() string
	// headers returns the HTTP headers required for authentication and API versioning.
	headers(apiKey string) internalhttp.Headers
	// buildRequestBody constructs the provider-specific request payload from a ChatRequest.
	buildRequestBody(req ChatRequest) (any, error)
	// parseResponse parses the provider's JSON response body into a ChatResponse.
	parseResponse(body string) (*ChatResponse, error)
}

// Provider implementations
var providers = map[string]provider{
	"openai":    openAIProvider{},
	"anthropic": anthropicProvider{},
}

const anthropicVersion = "2023-06-01"

// Provider environment variable names
const (
	openAIAPIKeyEnv      = "OPENAI_API_KEY"
	openAIEndpointEnv    = "OPENAI_ENDPOINT"
	anthropicAPIKeyEnv   = "ANTHROPIC_API_KEY"
	anthropicEndpointEnv = "ANTHROPIC_ENDPOINT"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a unified chat request
type ChatRequest struct {
	Provider    string
	Model       string
	Messages    []Message
	MaxTokens   int
	Temperature float64
	Endpoint    string // Optional custom endpoint URL (overrides env)
}

// ChatResponse represents the unified response from AI providers
type ChatResponse struct {
	Content string
	Model   string
	Usage   Usage
	// Tracking info for logging/debugging
	Endpoint     string // Full URL used for the request
	RequestJSON  string // Raw request body JSON
	ResponseJSON string // Raw response body JSON
}

// Usage contains token usage information
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Client is an interface for making AI chat requests
type Client interface {
	Chat(functionID string, req ChatRequest) (*ChatResponse, error)
}

// DefaultClient is the default implementation of Client
type DefaultClient struct {
	httpClient internalhttp.Client
	envStore   env.Store
}

// NewDefaultClient creates a new AI client
func NewDefaultClient(httpClient internalhttp.Client, envStore env.Store) *DefaultClient {
	return &DefaultClient{
		httpClient: httpClient,
		envStore:   envStore,
	}
}

// Chat executes a chat request using the specified provider
func (c *DefaultClient) Chat(functionID string, req ChatRequest) (*ChatResponse, error) {
	p, ok := providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s (use openai or anthropic)", req.Provider)
	}

	// Get API key and endpoint from environment
	apiKey, endpoint, err := c.getProviderConfig(functionID, req.Provider)
	if err != nil {
		return nil, err
	}

	// Allow endpoint override from request
	if req.Endpoint != "" {
		endpoint = req.Endpoint
	}

	return c.callProvider(req, p, apiKey, endpoint)
}

// getProviderConfig returns the API key and endpoint for the given provider
func (c *DefaultClient) getProviderConfig(functionID, providerName string) (apiKey, endpoint string, err error) {
	switch providerName {
	case "openai":
		apiKey, err = c.envStore.Get(functionID, openAIAPIKeyEnv)
		if err != nil || apiKey == "" {
			return "", "", fmt.Errorf("%s not set in function environment", openAIAPIKeyEnv)
		}
		endpoint, _ = c.envStore.Get(functionID, openAIEndpointEnv)
	case "anthropic":
		apiKey, err = c.envStore.Get(functionID, anthropicAPIKeyEnv)
		if err != nil || apiKey == "" {
			return "", "", fmt.Errorf("%s not set in function environment", anthropicAPIKeyEnv)
		}
		endpoint, _ = c.envStore.Get(functionID, anthropicEndpointEnv)
	default:
		return "", "", fmt.Errorf("unsupported provider: %s (use openai or anthropic)", providerName)
	}
	return apiKey, endpoint, nil
}

// callProvider executes an AI request using the given provider
func (c *DefaultClient) callProvider(req ChatRequest, p provider, apiKey, endpoint string) (*ChatResponse, error) {
	// Determine endpoint
	if endpoint == "" {
		endpoint = p.defaultEndpoint()
	}
	fullURL := endpoint + p.urlPath()

	// Build request body
	reqBody, err := p.buildRequestBody(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create partial response with tracking info (returned even on error)
	chatResp := &ChatResponse{
		Endpoint:    fullURL,
		RequestJSON: string(jsonBody),
	}

	// Make HTTP request
	resp, err := c.httpClient.Post(internalhttp.Request{
		URL:     fullURL,
		Headers: p.headers(apiKey),
		Body:    string(jsonBody),
	})
	if err != nil {
		return chatResp, fmt.Errorf("HTTP request failed: %v", err)
	}

	// Store response body for tracking
	chatResp.ResponseJSON = resp.Body

	// Parse response
	parsedResp, err := p.parseResponse(resp.Body)
	if err != nil {
		return chatResp, err
	}

	// Copy parsed response fields
	chatResp.Content = parsedResp.Content
	chatResp.Model = parsedResp.Model
	chatResp.Usage = parsedResp.Usage

	return chatResp, nil
}
