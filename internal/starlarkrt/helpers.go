package starlarkrt

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// builtinFn is the signature every host function implements.
type builtinFn = func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)

// module builds a named Starlark module from a set of members.
func module(name string, members starlark.StringDict) *starlarkstruct.Module {
	return &starlarkstruct.Module{Name: name, Members: members}
}

// builtin wraps a Go function as a named Starlark built-in.
func builtin(name string, fn builtinFn) *starlark.Builtin {
	return starlark.NewBuiltin(name, fn)
}

// okResult is the success arm of the (value, error) tuple convention.
func okResult(v starlark.Value) starlark.Value {
	return starlark.Tuple{v, starlark.None}
}

// errResult is the failure arm of the (value, error) tuple convention.
func errResult(msg string) starlark.Value {
	return starlark.Tuple{starlark.None, starlark.String(msg)}
}

// optDict returns the named entry of d as a dict, or nil when absent or not a dict.
func optDict(d *starlark.Dict, key string) *starlark.Dict {
	if d == nil {
		return nil
	}
	v, found, _ := d.Get(starlark.String(key))
	if !found {
		return nil
	}
	nested, _ := v.(*starlark.Dict)
	return nested
}

// optString returns the named entry of d coerced to a string, or "" when absent.
func optString(d *starlark.Dict, key string) string {
	if d == nil {
		return ""
	}
	v, found, _ := d.Get(starlark.String(key))
	if !found {
		return ""
	}
	return asString(v)
}
