package ai

import (
	"encoding/json"
	"fmt"

	internalhttp "github.com/dimiro1/faas-go/internal/http"
)

// anthropicProvider implements provider for Anthropic
type anthropicProvider struct{}

func (anthropicProvider) defaultEndpoint() string { return "https://api.anthropic.com" }
func (anthropicProvider) urlPath() string         { return "/v1/messages" }

func (anthropicProvider) headers(apiKey string) internalhttp.Headers {
	return internalhttp.Headers{
		"Content-Type":      "application/json",
		"x-api-key":         apiKey,
		"anthropic-version": anthropicVersion,
	}
}

func (anthropicProvider) buildRequestBody(req ChatRequest) (any, error) {
	// Extract system message (Anthropic handles it separately)
	var systemPrompt string
	var userMessages []Message
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			userMessages = append(userMessages, msg)
		}
	}

	body := struct {
		Model       string    `json:"model"`
		MaxTokens   int       `json:"max_tokens"`
		System      string    `json:"system,omitempty"`
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature,omitempty"`
	}{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		System:    systemPrompt,
		Messages:  userMessages,
	}
	if req.Temperature > 0 {
		body.Temperature = req.Temperature
	}
	return body, nil
}

func (anthropicProvider) parseResponse(body string) (*ChatResponse, error) {
	var resp struct {
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("anthropic API error: %s", resp.Error.Message)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	// Concatenate all text content
	var content string
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &ChatResponse{
		Content: content,
		Model:   resp.Model,
		Usage: Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
	}, nil
}
