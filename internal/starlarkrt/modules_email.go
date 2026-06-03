package starlarkrt

import (
	"time"

	stdlibemail "github.com/dimiro1/lunar/internal/runtime/email"
	"github.com/dimiro1/lunar/internal/services/email"
	"go.starlark.net/starlark"
)

// emailModule exposes email.send(options), returning a (result, error) tuple
// where result is {"id": "..."}. The tracked client records each send.
func emailModule(client email.Client, functionID string, tracker email.Tracker, executionID string) starlark.Value {
	tracked := stdlibemail.NewTrackedClient(client, tracker, executionID)

	return module("email", starlark.StringDict{
		"send": builtin("email.send", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var options *starlark.Dict
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "options", &options); err != nil {
				return nil, err
			}

			req, errMsg := parseEmailSendRequest(options)
			if errMsg != "" {
				return errResult(errMsg), nil
			}
			if err := stdlibemail.ValidateSendRequest(req); err != nil {
				return errResult(err.Error()), nil
			}

			res := tracked.SendWithTracking(functionID, req)
			if res.Error != nil {
				return errResult(res.Error.Error()), nil
			}

			result := starlark.NewDict(1)
			_ = result.SetKey(starlark.String("id"), starlark.String(res.Response.ID))
			return okResult(result), nil
		}),
	})
}

// parseEmailSendRequest extracts an email.SendRequest from the options dict.
func parseEmailSendRequest(options *starlark.Dict) (email.SendRequest, string) {
	toVal, found, _ := options.Get(starlark.String("to"))
	if !found {
		return email.SendRequest{}, "to is required"
	}
	to := valueToStringSlice(toVal)
	if len(to) == 0 {
		return email.SendRequest{}, "to cannot be empty"
	}

	var scheduledAt string
	if sa, found, _ := options.Get(starlark.String("scheduled_at")); found {
		switch v := sa.(type) {
		case starlark.Int:
			if ts, ok := v.Int64(); ok {
				scheduledAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
			}
		case starlark.String:
			scheduledAt = string(v)
		}
	}

	var headers map[string]string
	if h := optDict(options, "headers"); h != nil {
		headers = dictToStringMap(h)
	}

	return email.SendRequest{
		From:        optString(options, "from"),
		To:          to,
		Subject:     optString(options, "subject"),
		Text:        optString(options, "text"),
		HTML:        optString(options, "html"),
		ReplyTo:     optString(options, "reply_to"),
		Cc:          optStringSlice(options, "cc"),
		Bcc:         optStringSlice(options, "bcc"),
		Headers:     headers,
		Tags:        parseEmailTags(options),
		ScheduledAt: scheduledAt,
	}, ""
}

// parseEmailTags reads the optional tags list ([{name, value}, ...]).
func parseEmailTags(options *starlark.Dict) []email.Tag {
	v, found, _ := options.Get(starlark.String("tags"))
	if !found {
		return nil
	}
	iterable, ok := v.(starlark.Iterable)
	if !ok {
		return nil
	}
	iter := iterable.Iterate()
	defer iter.Done()

	var tags []email.Tag
	var elem starlark.Value
	for iter.Next(&elem) {
		d, ok := elem.(*starlark.Dict)
		if !ok {
			continue
		}
		tag := email.Tag{Name: optString(d, "name"), Value: optString(d, "value")}
		if tag.Name != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// optStringSlice reads a field that may be a string or a list of strings.
func optStringSlice(options *starlark.Dict, key string) []string {
	v, found, _ := options.Get(starlark.String(key))
	if !found {
		return nil
	}
	return valueToStringSlice(v)
}

// valueToStringSlice coerces a string or an iterable of strings to a slice,
// dropping empty entries.
func valueToStringSlice(v starlark.Value) []string {
	if s, ok := starlark.AsString(v); ok {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	iterable, ok := v.(starlark.Iterable)
	if !ok {
		return nil
	}
	iter := iterable.Iterate()
	defer iter.Done()

	var out []string
	var elem starlark.Value
	for iter.Next(&elem) {
		if s := asString(elem); s != "" {
			out = append(out, s)
		}
	}
	return out
}
