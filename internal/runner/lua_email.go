package runner

import (
	"time"

	"github.com/dimiro1/lunar/internal/services/email"
	stdlibemail "github.com/dimiro1/lunar/internal/runtime/email"
	lua "github.com/yuin/gopher-lua"
)

// registerEmail creates the global 'email' table with email sending functions.
// This is a thin wrapper using the stdlib/email TrackedClient decorator.
func registerEmail(L *lua.LState, emailClient email.Client, functionID string, emailTracker email.Tracker, executionID string) {
	// Create tracked client (decorator pattern)
	trackedClient := stdlibemail.NewTrackedClient(emailClient, emailTracker, executionID)

	emailTable := L.NewTable()

	// email.send(options)
	L.SetField(emailTable, "send", L.NewFunction(func(L *lua.LState) int {
		options := L.CheckTable(1)

		// Parse request from Lua options
		req, err := parseEmailSendRequest(options)
		if err != "" {
			L.Push(lua.LNil)
			L.Push(lua.LString(err))
			return 2
		}

		// Validate using reusable validation
		if validationErr := stdlibemail.ValidateSendRequest(req); validationErr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(validationErr.Error()))
			return 2
		}

		// Send with automatic tracking via decorator
		result := trackedClient.SendWithTracking(functionID, req)

		if result.Error != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(result.Error.Error()))
			return 2
		}

		// Convert response to Lua table
		resultTbl := L.NewTable()
		L.SetField(resultTbl, "id", lua.LString(result.Response.ID))
		L.Push(resultTbl)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("email", emailTable)
}

// parseEmailSendRequest extracts email.SendRequest from Lua options table
func parseEmailSendRequest(options *lua.LTable) (email.SendRequest, string) {
	from := lua.LVAsString(options.RawGetString("from"))
	toLV := options.RawGetString("to")
	subject := lua.LVAsString(options.RawGetString("subject"))
	text := lua.LVAsString(options.RawGetString("text"))
	html := lua.LVAsString(options.RawGetString("html"))
	replyTo := lua.LVAsString(options.RawGetString("reply_to"))

	// Convert 'to' to slice - can be string or table
	var to []string
	switch toLV.Type() {
	case lua.LTString:
		to = []string{lua.LVAsString(toLV)}
	case lua.LTTable:
		to = luaTableToStringSlice(toLV.(*lua.LTable))
		if len(to) == 0 {
			return email.SendRequest{}, "to cannot be empty"
		}
	case lua.LTNil:
		return email.SendRequest{}, "to is required"
	default:
		return email.SendRequest{}, "to must be a string or table of strings"
	}

	// Handle scheduled_at - accepts Unix timestamp (number) or ISO 8601 string
	var scheduledAt string
	scheduledAtLV := options.RawGetString("scheduled_at")
	switch scheduledAtLV.Type() {
	case lua.LTNumber:
		// Convert Unix timestamp to ISO 8601
		ts := int64(lua.LVAsNumber(scheduledAtLV))
		scheduledAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
	case lua.LTString:
		scheduledAt = lua.LVAsString(scheduledAtLV)
	}

	// Convert optional cc and bcc
	cc := extractStringSliceFromLua(options.RawGetString("cc"))
	bcc := extractStringSliceFromLua(options.RawGetString("bcc"))

	// Convert optional headers
	var headers map[string]string
	headersLV := options.RawGetString("headers")
	if headersLV.Type() == lua.LTTable {
		headers = make(map[string]string)
		headersLV.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			headers[lua.LVAsString(k)] = lua.LVAsString(v)
		})
	}

	// Convert optional tags
	var tags []email.Tag
	tagsLV := options.RawGetString("tags")
	if tagsLV.Type() == lua.LTTable {
		tagsLV.(*lua.LTable).ForEach(func(_, v lua.LValue) {
			if tagTbl, ok := v.(*lua.LTable); ok {
				tag := email.Tag{
					Name:  lua.LVAsString(tagTbl.RawGetString("name")),
					Value: lua.LVAsString(tagTbl.RawGetString("value")),
				}
				if tag.Name != "" {
					tags = append(tags, tag)
				}
			}
		})
	}

	return email.SendRequest{
		From:        from,
		To:          to,
		Subject:     subject,
		Text:        text,
		HTML:        html,
		ReplyTo:     replyTo,
		Cc:          cc,
		Bcc:         bcc,
		Headers:     headers,
		Tags:        tags,
		ScheduledAt: scheduledAt,
	}, ""
}

// extractStringSliceFromLua extracts a string slice from a Lua value
func extractStringSliceFromLua(lv lua.LValue) []string {
	switch lv.Type() {
	case lua.LTString:
		return []string{lua.LVAsString(lv)}
	case lua.LTTable:
		return luaTableToStringSlice(lv.(*lua.LTable))
	default:
		return nil
	}
}

// luaTableToStringSlice converts a Lua table to a slice of strings
func luaTableToStringSlice(tbl *lua.LTable) []string {
	var result []string
	tbl.ForEach(func(_, v lua.LValue) {
		if str := lua.LVAsString(v); str != "" {
			result = append(result, str)
		}
	})
	return result
}
