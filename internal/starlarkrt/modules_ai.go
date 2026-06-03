package starlarkrt

import (
	stdlibai "github.com/dimiro1/lunar/internal/runtime/ai"
	"github.com/dimiro1/lunar/internal/services/ai"
	"go.starlark.net/starlark"
)

// aiModule exposes ai.chat(options), returning a (response, error) tuple. The
// tracked client records each call against the execution, as in the Lua runtime.
func aiModule(client ai.Client, functionID string, tracker ai.Tracker, executionID string) starlark.Value {
	tracked := stdlibai.NewTrackedClient(client, tracker, executionID)

	return module("ai", starlark.StringDict{
		"chat": builtin("ai.chat", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var options *starlark.Dict
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "options", &options); err != nil {
				return nil, err
			}

			req, errMsg := parseAIChatRequest(options)
			if errMsg != "" {
				return errResult(errMsg), nil
			}

			res := tracked.ChatWithTracking(functionID, req)
			if res.Error != nil {
				return errResult(res.Error.Error()), nil
			}
			return okResult(aiResponseToStarlark(res.Response)), nil
		}),
	})
}

// parseAIChatRequest extracts an ai.ChatRequest from the options dict, returning
// a human-readable message when a required field is missing or malformed.
func parseAIChatRequest(options *starlark.Dict) (ai.ChatRequest, string) {
	provider := optString(options, "provider")
	model := optString(options, "model")

	if provider == "" {
		return ai.ChatRequest{}, "provider is required (openai or anthropic)"
	}
	if model == "" {
		return ai.ChatRequest{}, "model is required"
	}

	messagesVal, found, _ := options.Get(starlark.String("messages"))
	if !found {
		return ai.ChatRequest{}, "messages is required and must be a list"
	}
	messages := parseAIMessages(messagesVal)
	if len(messages) == 0 {
		return ai.ChatRequest{}, "messages cannot be empty"
	}

	maxTokens := 1024
	if mt, found, _ := options.Get(starlark.String("max_tokens")); found {
		var n int
		if err := starlark.AsInt(mt, &n); err == nil && n > 0 {
			maxTokens = n
		}
	}

	var temperature float64
	if t, found, _ := options.Get(starlark.String("temperature")); found {
		temperature, _ = starlark.AsFloat(t)
	}

	return ai.ChatRequest{
		Provider:    provider,
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Endpoint:    optString(options, "endpoint"),
	}, ""
}

// parseAIMessages converts the messages iterable to a slice of ai.Message,
// skipping entries that lack a role or content.
func parseAIMessages(v starlark.Value) []ai.Message {
	iterable, ok := v.(starlark.Iterable)
	if !ok {
		return nil
	}
	iter := iterable.Iterate()
	defer iter.Done()

	var messages []ai.Message
	var elem starlark.Value
	for iter.Next(&elem) {
		d, ok := elem.(*starlark.Dict)
		if !ok {
			continue
		}
		msg := ai.Message{
			Role:    optString(d, "role"),
			Content: optString(d, "content"),
		}
		if msg.Role != "" && msg.Content != "" {
			messages = append(messages, msg)
		}
	}
	return messages
}

// aiResponseToStarlark converts an AI chat response to a Starlark dict.
func aiResponseToStarlark(resp *ai.ChatResponse) starlark.Value {
	usage := starlark.NewDict(2)
	_ = usage.SetKey(starlark.String("input_tokens"), starlark.MakeInt(resp.Usage.InputTokens))
	_ = usage.SetKey(starlark.String("output_tokens"), starlark.MakeInt(resp.Usage.OutputTokens))

	d := starlark.NewDict(3)
	_ = d.SetKey(starlark.String("content"), starlark.String(resp.Content))
	_ = d.SetKey(starlark.String("model"), starlark.String(resp.Model))
	_ = d.SetKey(starlark.String("usage"), usage)
	return d
}
