package starlarkrt

import (
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"go.starlark.net/starlark"
)

// httpModule exposes the HTTP client: http.get/post/put/delete(url, options).
// Fallible calls return a (response, error) tuple, e.g.:
//
//	resp, err = http.get("https://example.com", {"headers": {"Accept": "application/json"}})
func httpModule(client internalhttp.Client) starlark.Value {
	verb := func(name string, withBody bool, do func(internalhttp.Request) (internalhttp.Response, error)) *starlark.Builtin {
		return builtin("http."+name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var url string
			var options starlark.Value = starlark.None
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &url, "options?", &options); err != nil {
				return nil, err
			}

			req := internalhttp.Request{URL: url}
			if opts, ok := options.(*starlark.Dict); ok {
				if headers := optDict(opts, "headers"); headers != nil {
					req.Headers = internalhttp.Headers(dictToStringMap(headers))
				}
				if query := optDict(opts, "query"); query != nil {
					req.Query = internalhttp.Query(dictToStringMap(query))
				}
				if withBody {
					req.Body = optString(opts, "body")
				}
			}

			resp, err := do(req)
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(httpResponseToStarlark(resp)), nil
		})
	}

	return module("http", starlark.StringDict{
		"get":    verb("get", false, client.Get),
		"post":   verb("post", true, client.Post),
		"put":    verb("put", true, client.Put),
		"delete": verb("delete", false, client.Delete),
	})
}

// httpResponseToStarlark converts an HTTP client response to a Starlark dict.
func httpResponseToStarlark(resp internalhttp.Response) starlark.Value {
	d := starlark.NewDict(3)
	_ = d.SetKey(starlark.String("statusCode"), starlark.MakeInt(resp.StatusCode))
	_ = d.SetKey(starlark.String("body"), starlark.String(resp.Body))
	_ = d.SetKey(starlark.String("headers"), stringMapToDict(resp.Headers))
	return d
}
