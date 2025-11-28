package runner

import (
	"time"

	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/store"
	lua "github.com/yuin/gopher-lua"
)

// registerAI creates the global 'ai' table with AI provider functions
func registerAI(L *lua.LState, client ai.Client, functionID string, tracker ai.Tracker, executionID string) {
	aiTable := L.NewTable()

	// ai.chat(options)
	L.SetField(aiTable, "chat", L.NewFunction(func(L *lua.LState) int {
		options := L.CheckTable(1)

		// Extract required parameters
		provider := lua.LVAsString(options.RawGetString("provider"))
		model := lua.LVAsString(options.RawGetString("model"))
		messagesLV := options.RawGetString("messages")

		// Validate required parameters
		if provider == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("provider is required (openai or anthropic)"))
			return 2
		}
		if model == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("model is required"))
			return 2
		}
		if messagesLV.Type() != lua.LTTable {
			L.Push(lua.LNil)
			L.Push(lua.LString("messages is required and must be a table"))
			return 2
		}

		// Convert messages from Lua to Go
		messages := luaMessagesToGo(messagesLV.(*lua.LTable))
		if len(messages) == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("messages cannot be empty"))
			return 2
		}

		// Extract optional parameters
		maxTokens := int(lua.LVAsNumber(options.RawGetString("max_tokens")))
		temperature := lua.LVAsNumber(options.RawGetString("temperature"))
		endpoint := lua.LVAsString(options.RawGetString("endpoint"))

		// Set defaults for optional parameters
		if maxTokens == 0 {
			maxTokens = 1024
		}

		// Build chat request
		req := ai.ChatRequest{
			Provider:    provider,
			Model:       model,
			Messages:    messages,
			MaxTokens:   maxTokens,
			Temperature: float64(temperature),
			Endpoint:    endpoint,
		}

		// Execute the request with tracking
		response, trackReq := executeWithTracking(client, functionID, req)

		// Track the request (success or error)
		if tracker != nil {
			tracker.Track(executionID, trackReq)
		}

		if trackReq.Status == store.AIRequestStatusError {
			L.Push(lua.LNil)
			L.Push(lua.LString(*trackReq.ErrorMessage))
			return 2
		}

		// Convert response to Lua table
		L.Push(aiResponseToLuaTable(L, response))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("ai", aiTable)
}

// executeWithTracking executes an AI chat request and returns tracking info
func executeWithTracking(client ai.Client, functionID string, req ai.ChatRequest) (*ai.ChatResponse, ai.TrackRequest) {
	trackReq := ai.TrackRequest{
		Provider: req.Provider,
		Model:    req.Model,
	}

	startTime := time.Now()
	response, err := client.Chat(functionID, req)
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
		return nil, trackReq
	}

	trackReq.Status = store.AIRequestStatusSuccess
	trackReq.InputTokens = &response.Usage.InputTokens
	trackReq.OutputTokens = &response.Usage.OutputTokens

	return response, trackReq
}

// luaMessagesToGo converts a Lua table of messages to Go
func luaMessagesToGo(tbl *lua.LTable) []ai.Message {
	var messages []ai.Message
	tbl.ForEach(func(_, v lua.LValue) {
		if msgTbl, ok := v.(*lua.LTable); ok {
			msg := ai.Message{
				Role:    lua.LVAsString(msgTbl.RawGetString("role")),
				Content: lua.LVAsString(msgTbl.RawGetString("content")),
			}
			if msg.Role != "" && msg.Content != "" {
				messages = append(messages, msg)
			}
		}
	})
	return messages
}

// aiResponseToLuaTable converts an AI response to a Lua table
func aiResponseToLuaTable(L *lua.LState, resp *ai.ChatResponse) *lua.LTable {
	tbl := L.NewTable()
	L.SetField(tbl, "content", lua.LString(resp.Content))
	L.SetField(tbl, "model", lua.LString(resp.Model))

	usageTbl := L.NewTable()
	L.SetField(usageTbl, "input_tokens", lua.LNumber(resp.Usage.InputTokens))
	L.SetField(usageTbl, "output_tokens", lua.LNumber(resp.Usage.OutputTokens))
	L.SetField(tbl, "usage", usageTbl)

	return tbl
}
