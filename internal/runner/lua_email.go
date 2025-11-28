package runner

import (
	"encoding/json"
	"time"

	"github.com/dimiro1/faas-go/internal/email"
	"github.com/dimiro1/faas-go/internal/store"
	lua "github.com/yuin/gopher-lua"
)

// registerEmail creates the global 'email' table with email sending functions
func registerEmail(L *lua.LState, emailClient email.Client, functionID string, emailTracker email.Tracker, executionID string) {
	emailTable := L.NewTable()

	// email.send(options)
	L.SetField(emailTable, "send", L.NewFunction(func(L *lua.LState) int {
		options := L.CheckTable(1)

		// Extract required parameters
		from := lua.LVAsString(options.RawGetString("from"))
		toLV := options.RawGetString("to")
		subject := lua.LVAsString(options.RawGetString("subject"))

		// Validate required parameters
		if from == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("from is required"))
			return 2
		}
		if toLV.Type() == lua.LTNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("to is required"))
			return 2
		}
		if subject == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("subject is required"))
			return 2
		}

		// Convert 'to' to slice - can be string or table
		var to []string
		if toLV.Type() == lua.LTString {
			to = []string{lua.LVAsString(toLV)}
		} else if toLV.Type() == lua.LTTable {
			to = luaTableToStringSlice(toLV.(*lua.LTable))
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LString("to must be a string or table of strings"))
			return 2
		}

		if len(to) == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("to cannot be empty"))
			return 2
		}

		// Extract optional parameters
		text := lua.LVAsString(options.RawGetString("text"))
		html := lua.LVAsString(options.RawGetString("html"))
		replyTo := lua.LVAsString(options.RawGetString("reply_to"))

		// Handle scheduled_at - accepts Unix timestamp (number) or ISO 8601 string
		var scheduledAt string
		scheduledAtLV := options.RawGetString("scheduled_at")
		if scheduledAtLV.Type() == lua.LTNumber {
			// Convert Unix timestamp to ISO 8601
			ts := int64(lua.LVAsNumber(scheduledAtLV))
			scheduledAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		} else if scheduledAtLV.Type() == lua.LTString {
			scheduledAt = lua.LVAsString(scheduledAtLV)
		}

		// At least text or html must be provided
		if text == "" && html == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("either text or html content is required"))
			return 2
		}

		// Convert optional cc and bcc
		var cc, bcc []string
		ccLV := options.RawGetString("cc")
		if ccLV.Type() == lua.LTString {
			cc = []string{lua.LVAsString(ccLV)}
		} else if ccLV.Type() == lua.LTTable {
			cc = luaTableToStringSlice(ccLV.(*lua.LTable))
		}

		bccLV := options.RawGetString("bcc")
		if bccLV.Type() == lua.LTString {
			bcc = []string{lua.LVAsString(bccLV)}
		} else if bccLV.Type() == lua.LTTable {
			bcc = luaTableToStringSlice(bccLV.(*lua.LTable))
		}

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

		// Build the send request
		req := email.SendRequest{
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
		}

		// Send the email
		startTime := time.Now()
		resp, err := emailClient.Send(functionID, req)
		durationMs := time.Since(startTime).Milliseconds()

		// Get request JSON for tracking (available even on error)
		var requestJSON string
		if resp != nil {
			requestJSON = resp.RequestJSON
		}

		if err != nil {
			// Track failed request
			errMsg := err.Error()
			if emailTracker != nil {
				emailTracker.Track(executionID, email.TrackRequest{
					From:         from,
					To:           to,
					Subject:      subject,
					HasText:      text != "",
					HasHTML:      html != "",
					RequestJSON:  requestJSON,
					Status:       store.EmailRequestStatusError,
					ErrorMessage: &errMsg,
					DurationMs:   durationMs,
				})
			}
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Build response JSON
		responseJSON, _ := json.Marshal(map[string]string{"id": resp.ID})
		responseJSONStr := string(responseJSON)

		// Track successful request
		if emailTracker != nil {
			emailTracker.Track(executionID, email.TrackRequest{
				From:         from,
				To:           to,
				Subject:      subject,
				HasText:      text != "",
				HasHTML:      html != "",
				RequestJSON:  requestJSON,
				ResponseJSON: &responseJSONStr,
				Status:       store.EmailRequestStatusSuccess,
				EmailID:      &resp.ID,
				DurationMs:   durationMs,
			})
		}

		// Convert response to Lua table
		result := L.NewTable()
		L.SetField(result, "id", lua.LString(resp.ID))

		L.Push(result)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("email", emailTable)
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
