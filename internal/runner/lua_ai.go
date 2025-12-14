package runner

import (
	"github.com/dimiro1/lunar/internal/services/ai"
	stdlibai "github.com/dimiro1/lunar/internal/runtime/ai"
	lua "github.com/yuin/gopher-lua"
)

// registerAI creates the global 'ai' table with AI provider functions.
// This is a thin wrapper using the stdlib/ai TrackedClient decorator.
func registerAI(L *lua.LState, client ai.Client, functionID string, tracker ai.Tracker, executionID string) {
	trackedClient := stdlibai.NewTrackedClient(client, tracker, executionID)

	aiTable := L.NewTable()

	// ai.chat(options)
	L.SetField(aiTable, "chat", L.NewFunction(func(L *lua.LState) int {
		options := L.CheckTable(1)

		// Extract and validate parameters
		req, errMsg := parseAIChatRequest(options)
		if errMsg != "" {
			L.Push(lua.LNil)
			L.Push(lua.LString(errMsg))
			return 2
		}

		// Execute with automatic tracking via decorator
		result := trackedClient.ChatWithTracking(functionID, req)

		if result.Error != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(result.Error.Error()))
			return 2
		}

		L.Push(aiResponseToLuaTable(L, result.Response))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("ai", aiTable)
}

// parseAIChatRequest extracts ai.ChatRequest from Lua options table
func parseAIChatRequest(options *lua.LTable) (ai.ChatRequest, string) {
	provider := lua.LVAsString(options.RawGetString("provider"))
	model := lua.LVAsString(options.RawGetString("model"))
	messagesLV := options.RawGetString("messages")

	// Validate required parameters
	if provider == "" {
		return ai.ChatRequest{}, "provider is required (openai or anthropic)"
	}
	if model == "" {
		return ai.ChatRequest{}, "model is required"
	}
	if messagesLV.Type() != lua.LTTable {
		return ai.ChatRequest{}, "messages is required and must be a table"
	}

	// Convert messages from Lua to Go
	messages := luaMessagesToGo(messagesLV.(*lua.LTable))
	if len(messages) == 0 {
		return ai.ChatRequest{}, "messages cannot be empty"
	}

	// Extract optional parameters
	maxTokens := int(lua.LVAsNumber(options.RawGetString("max_tokens")))
	temperature := lua.LVAsNumber(options.RawGetString("temperature"))
	endpoint := lua.LVAsString(options.RawGetString("endpoint"))

	// Set defaults for optional parameters
	if maxTokens == 0 {
		maxTokens = 1024
	}

	return ai.ChatRequest{
		Provider:    provider,
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float64(temperature),
		Endpoint:    endpoint,
	}, ""
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
