package starlarkrt

import (
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/runtime/router"
	"go.starlark.net/starlark"
)

// routerModule exposes path matching and per-function path/URL building.
func routerModule(ctx *events.ExecutionContext) starlark.Value {
	return module("router", starlark.StringDict{
		"match": builtin("router.match", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path, pattern string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path, "pattern", &pattern); err != nil {
				return nil, err
			}
			return starlark.Bool(router.Match(path, pattern).Matched), nil
		}),
		"params": builtin("router.params", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path, pattern string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path, "pattern", &pattern); err != nil {
				return nil, err
			}
			res := router.Match(path, pattern)
			if !res.Matched {
				return starlark.NewDict(0), nil
			}
			return stringMapToDict(res.Params), nil
		}),
		"path": builtin("router.path", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			pattern, params, err := unpackPattern(b, args, kwargs)
			if err != nil {
				return nil, err
			}
			return starlark.String(router.FunctionPath(ctx.FunctionID, pattern, params)), nil
		}),
		"url": builtin("router.url", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			pattern, params, err := unpackPattern(b, args, kwargs)
			if err != nil {
				return nil, err
			}
			return starlark.String(router.FunctionURL(ctx.BaseURL, ctx.FunctionID, pattern, params)), nil
		}),
	})
}

// unpackPattern reads a required pattern and an optional params dict.
func unpackPattern(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, map[string]string, error) {
	var pattern string
	var params starlark.Value = starlark.None
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pattern, "params?", &params); err != nil {
		return "", nil, err
	}
	if d, ok := params.(*starlark.Dict); ok {
		return pattern, dictToStringMap(d), nil
	}
	return pattern, nil, nil
}
