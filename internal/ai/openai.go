package ai

import (
	"encoding/json"
	"fmt"

	internalhttp "github.com/dimiro1/faas-go/internal/http"
)

// openAIProvider implements provider for OpenAI
type openAIProvider struct{}

func (openAIProvider) defaultEndpoint() string { return "https://api.openai.com/v1" }
func (openAIProvider) urlPath() string         { return "/chat/completions" }

func (openAIProvider) headers(apiKey string) internalhttp.Headers {
	return internalhttp.Headers{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}
}

func (openAIProvider) buildRequestBody(req ChatRequest) (any, error) {
	body := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		MaxTokens   int       `json:"max_tokens,omitempty"`
		Temperature float64   `json:"temperature,omitempty"`
	}{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: req.MaxTokens,
	}
	if req.Temperature > 0 {
		body.Temperature = req.Temperature
	}
	return body, nil
}

func (openAIProvider) parseResponse(body string) (*ChatResponse, error) {
	var resp struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &ChatResponse{
		Content: resp.Choices[0].Message.Content,
		Model:   resp.Model,
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}
