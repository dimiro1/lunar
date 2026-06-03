package starlarkrt

import (
	"github.com/dimiro1/lunar/internal/events"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// contextToStarlark converts an ExecutionContext to a Starlark struct so handler
// code can use attribute access: ctx.executionId, ctx.functionId, etc.
func contextToStarlark(ctx *events.ExecutionContext) starlark.Value {
	d := starlark.StringDict{
		"executionId": starlark.String(ctx.ExecutionID),
		"functionId":  starlark.String(ctx.FunctionID),
		"startedAt":   starlark.MakeInt64(ctx.StartedAt),
	}
	if ctx.RequestID != "" {
		d["requestId"] = starlark.String(ctx.RequestID)
	}
	if ctx.FunctionName != "" {
		d["functionName"] = starlark.String(ctx.FunctionName)
	}
	if ctx.Version != "" {
		d["version"] = starlark.String(ctx.Version)
	}
	if ctx.BaseURL != "" {
		d["baseUrl"] = starlark.String(ctx.BaseURL)
	}
	return starlarkstruct.FromStringDict(starlarkstruct.Default, d)
}

// httpEventToStarlark converts an HTTPEvent to a Starlark struct.
func httpEventToStarlark(event events.HTTPEvent) starlark.Value {
	d := starlark.StringDict{
		"method":       starlark.String(event.Method),
		"path":         starlark.String(event.Path),
		"relativePath": starlark.String(event.RelativePath),
		"body":         starlark.String(event.Body),
		"headers":      stringMapToDict(event.Headers),
		"query":        stringMapToDict(event.Query),
	}
	return starlarkstruct.FromStringDict(starlarkstruct.Default, d)
}

// starlarkToHTTPResponse converts the handler's return value to an HTTPResponse.
// It accepts a dict (idiomatic) or any value exposing attributes (a struct).
func starlarkToHTTPResponse(v starlark.Value) events.HTTPResponse {
	resp := events.HTTPResponse{
		StatusCode: 200, // Default
		Headers:    make(map[string]string),
	}

	if sc, ok := fieldOf(v, "statusCode"); ok {
		var code int
		if err := starlark.AsInt(sc, &code); err == nil {
			resp.StatusCode = code
		}
	}

	if body, ok := fieldOf(v, "body"); ok {
		resp.Body = asString(body)
	}

	if headers, ok := fieldOf(v, "headers"); ok {
		if d, isDict := headers.(*starlark.Dict); isDict {
			for k, val := range dictToStringMap(d) {
				resp.Headers[k] = val
			}
		}
	}

	if b64, ok := fieldOf(v, "isBase64Encoded"); ok {
		resp.IsBase64Encoded = bool(b64.Truth())
	}

	return resp
}

// fieldOf reads a named field from a dict, a struct, or any value with attributes.
func fieldOf(v starlark.Value, name string) (starlark.Value, bool) {
	if d, ok := v.(*starlark.Dict); ok {
		val, found, _ := d.Get(starlark.String(name))
		return val, found
	}
	if ha, ok := v.(starlark.HasAttrs); ok {
		val, err := ha.Attr(name)
		if err == nil && val != nil {
			return val, true
		}
	}
	return nil, false
}

// stringMapToDict converts a Go string map to a Starlark dict.
func stringMapToDict(m map[string]string) *starlark.Dict {
	d := starlark.NewDict(len(m))
	for k, v := range m {
		_ = d.SetKey(starlark.String(k), starlark.String(v))
	}
	return d
}

// dictToStringMap converts a Starlark dict to a Go string map. Both keys and
// values are coerced to strings.
func dictToStringMap(d *starlark.Dict) map[string]string {
	out := make(map[string]string, d.Len())
	for _, item := range d.Items() {
		out[asString(item[0])] = asString(item[1])
	}
	return out
}

// asString coerces a Starlark value to a Go string. Strings are returned
// verbatim; other values use their Starlark representation (e.g. 200 -> "200").
func asString(v starlark.Value) string {
	if s, ok := starlark.AsString(v); ok {
		return s
	}
	return v.String()
}

// goToStarlark converts a decoded Go value (from JSON, etc.) to a Starlark value.
func goToStarlark(v any) starlark.Value {
	switch val := v.(type) {
	case nil:
		return starlark.None
	case bool:
		return starlark.Bool(val)
	case float64:
		return starlark.Float(val)
	case int:
		return starlark.MakeInt(val)
	case int64:
		return starlark.MakeInt64(val)
	case string:
		return starlark.String(val)
	case []any:
		elems := make([]starlark.Value, len(val))
		for i, e := range val {
			elems[i] = goToStarlark(e)
		}
		return starlark.NewList(elems)
	case map[string]any:
		d := starlark.NewDict(len(val))
		for k, e := range val {
			_ = d.SetKey(starlark.String(k), goToStarlark(e))
		}
		return d
	default:
		return starlark.None
	}
}

// starlarkToGo converts a Starlark value to a plain Go value for JSON encoding.
func starlarkToGo(v starlark.Value) any {
	switch val := v.(type) {
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return bool(val)
	case starlark.Int:
		if i, ok := val.Int64(); ok {
			return i
		}
		f, _ := starlark.AsFloat(val)
		return f
	case starlark.Float:
		return float64(val)
	case starlark.String:
		return string(val)
	case *starlark.List:
		out := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			out[i] = starlarkToGo(val.Index(i))
		}
		return out
	case starlark.Tuple:
		out := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			out[i] = starlarkToGo(val.Index(i))
		}
		return out
	case *starlark.Dict:
		out := make(map[string]any, val.Len())
		for _, item := range val.Items() {
			out[asString(item[0])] = starlarkToGo(item[1])
		}
		return out
	case *starlarkstruct.Struct:
		out := make(map[string]any)
		for _, name := range val.AttrNames() {
			if av, err := val.Attr(name); err == nil {
				out[name] = starlarkToGo(av)
			}
		}
		return out
	default:
		return nil
	}
}
