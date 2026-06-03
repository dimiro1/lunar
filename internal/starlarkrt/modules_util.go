package starlarkrt

import (
	"context"
	"fmt"

	stdlibbase64 "github.com/dimiro1/lunar/internal/runtime/base64"
	"github.com/dimiro1/lunar/internal/runtime/crypto"
	stdlibjson "github.com/dimiro1/lunar/internal/runtime/json"
	"github.com/dimiro1/lunar/internal/runtime/random"
	stdlibstrings "github.com/dimiro1/lunar/internal/runtime/strings"
	stdlibtime "github.com/dimiro1/lunar/internal/runtime/time"
	stdliburl "github.com/dimiro1/lunar/internal/runtime/url"
	"go.starlark.net/starlark"
)

// jsonModule exposes json.encode(value) and json.decode(str). Both return a
// (value, error) tuple following the Lua two-value convention.
func jsonModule() starlark.Value {
	return module("json", starlark.StringDict{
		"encode": builtin("json.encode", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var value starlark.Value
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
				return nil, err
			}
			str, err := stdlibjson.Encode(starlarkToGo(value))
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(starlark.String(str)), nil
		}),
		"decode": builtin("json.decode", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var str string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &str); err != nil {
				return nil, err
			}
			value, err := stdlibjson.Decode(str)
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(goToStarlark(value)), nil
		}),
	})
}

// base64Module exposes base64.encode(str) and base64.decode(str).
func base64Module() starlark.Value {
	return module("base64", starlark.StringDict{
		"encode": str1("base64.encode", func(s string) string { return stdlibbase64.Encode(s) }),
		"decode": builtin("base64.decode", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s); err != nil {
				return nil, err
			}
			decoded, err := stdlibbase64.Decode(s)
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(starlark.String(decoded)), nil
		}),
	})
}

// cryptoModule exposes hashing, HMAC and UUID helpers.
func cryptoModule() starlark.Value {
	return module("crypto", starlark.StringDict{
		"md5":         str1("crypto.md5", crypto.MD5),
		"sha1":        str1("crypto.sha1", crypto.SHA1),
		"sha256":      str1("crypto.sha256", crypto.SHA256),
		"sha512":      str1("crypto.sha512", crypto.SHA512),
		"hmac_sha1":   str2("crypto.hmac_sha1", crypto.HMACSHA1),
		"hmac_sha256": str2("crypto.hmac_sha256", crypto.HMACSHA256),
		"hmac_sha512": str2("crypto.hmac_sha512", crypto.HMACSHA512),
		"uuid": builtin("crypto.uuid", func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			return starlark.String(crypto.UUID()), nil
		}),
	})
}

// timeModule exposes time helpers. sleep honors the execution deadline via ctx.
func timeModule(ctx context.Context) starlark.Value {
	return module("time", starlark.StringDict{
		"now": builtin("time.now", func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			return starlark.MakeInt64(stdlibtime.Now()), nil
		}),
		"format": builtin("time.format", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var timestamp int64
			var layout string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "timestamp", &timestamp, "layout", &layout); err != nil {
				return nil, err
			}
			return starlark.String(stdlibtime.Format(timestamp, layout)), nil
		}),
		"parse": builtin("time.parse", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var timeStr, layout string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &timeStr, "layout", &layout); err != nil {
				return nil, err
			}
			ts, err := stdlibtime.Parse(timeStr, layout)
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(starlark.MakeInt64(ts)), nil
		}),
		"sleep": builtin("time.sleep", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var milliseconds int64
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "milliseconds", &milliseconds); err != nil {
				return nil, err
			}
			stdlibtime.Sleep(ctx, milliseconds)
			return starlark.None, nil
		}),
	})
}

// urlModule exposes URL parsing and percent-encoding.
func urlModule() starlark.Value {
	return module("url", starlark.StringDict{
		"parse": builtin("url.parse", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var urlStr string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &urlStr); err != nil {
				return nil, err
			}
			parsed, err := stdliburl.Parse(urlStr)
			if err != nil {
				return errResult(err.Error()), nil
			}

			d := starlark.NewDict(7)
			_ = d.SetKey(starlark.String("scheme"), starlark.String(parsed.Scheme))
			_ = d.SetKey(starlark.String("host"), starlark.String(parsed.Host))
			_ = d.SetKey(starlark.String("path"), starlark.String(parsed.Path))
			_ = d.SetKey(starlark.String("fragment"), starlark.String(parsed.Fragment))

			query := starlark.NewDict(len(parsed.Query))
			for key, values := range parsed.Query {
				if len(values) == 1 {
					_ = query.SetKey(starlark.String(key), starlark.String(values[0]))
				} else {
					elems := make([]starlark.Value, len(values))
					for i, v := range values {
						elems[i] = starlark.String(v)
					}
					_ = query.SetKey(starlark.String(key), starlark.NewList(elems))
				}
			}
			_ = d.SetKey(starlark.String("query"), query)

			if parsed.Username != "" {
				_ = d.SetKey(starlark.String("username"), starlark.String(parsed.Username))
			}
			if parsed.Password != "" {
				_ = d.SetKey(starlark.String("password"), starlark.String(parsed.Password))
			}
			return okResult(d), nil
		}),
		"encode": str1("url.encode", stdliburl.Encode),
		"decode": builtin("url.decode", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s); err != nil {
				return nil, err
			}
			decoded, err := stdliburl.Decode(s)
			if err != nil {
				return errResult(err.Error()), nil
			}
			return okResult(starlark.String(decoded)), nil
		}),
	})
}

// stringsModule exposes string manipulation helpers.
func stringsModule() starlark.Value {
	return module("strings", starlark.StringDict{
		"trim":      str1("strings.trim", stdlibstrings.Trim),
		"trimLeft":  str1("strings.trimLeft", stdlibstrings.TrimLeft),
		"trimRight": str1("strings.trimRight", stdlibstrings.TrimRight),
		"toLower":   str1("strings.toLower", stdlibstrings.ToLower),
		"toUpper":   str1("strings.toUpper", stdlibstrings.ToUpper),
		"hasPrefix": bool2("strings.hasPrefix", stdlibstrings.HasPrefix),
		"hasSuffix": bool2("strings.hasSuffix", stdlibstrings.HasSuffix),
		"contains":  bool2("strings.contains", stdlibstrings.Contains),
		"split": builtin("strings.split", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s, sep string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s, "sep", &sep); err != nil {
				return nil, err
			}
			parts := stdlibstrings.Split(s, sep)
			elems := make([]starlark.Value, len(parts))
			for i, p := range parts {
				elems[i] = starlark.String(p)
			}
			return starlark.NewList(elems), nil
		}),
		"join": builtin("strings.join", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var list starlark.Value
			var sep string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "list", &list, "sep", &sep); err != nil {
				return nil, err
			}
			return starlark.String(stdlibstrings.Join(iterableToStrings(list), sep)), nil
		}),
		"replace": builtin("strings.replace", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s, old, replacement string
			n := -1
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s, "old", &old, "new", &replacement, "n?", &n); err != nil {
				return nil, err
			}
			return starlark.String(stdlibstrings.Replace(s, old, replacement, n)), nil
		}),
		"repeat": builtin("strings.repeat", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s string
			var count int
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s, "count", &count); err != nil {
				return nil, err
			}
			return starlark.String(stdlibstrings.Repeat(s, count)), nil
		}),
	})
}

// randomModule exposes random generation helpers. int/string raise on invalid
// arguments; bytes/hex follow the (value, error) tuple convention.
func randomModule() starlark.Value {
	return module("random", starlark.StringDict{
		"int": builtin("random.int", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var minValue, maxValue int
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "min", &minValue, "max", &maxValue); err != nil {
				return nil, err
			}
			n, err := random.Int(minValue, maxValue)
			if err != nil {
				return nil, fmt.Errorf("random.int: %w", err)
			}
			return starlark.MakeInt(n), nil
		}),
		"float": builtin("random.float", func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			return starlark.Float(random.Float()), nil
		}),
		"string": builtin("random.string", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var length int
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "length", &length); err != nil {
				return nil, err
			}
			s, err := random.String(length)
			if err != nil {
				return nil, fmt.Errorf("random.string: %w", err)
			}
			return starlark.String(s), nil
		}),
		"bytes": lenToResult("random.bytes", random.Bytes),
		"hex":   lenToResult("random.hex", random.Hex),
		"id": builtin("random.id", func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			return starlark.String(random.ID()), nil
		}),
	})
}

// --- small builtin builders shared by the utility modules ---

// str1 builds a builtin taking one string and returning one string.
func str1(name string, fn func(string) string) *starlark.Builtin {
	return builtin(name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var s string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "str", &s); err != nil {
			return nil, err
		}
		return starlark.String(fn(s)), nil
	})
}

// str2 builds a builtin taking two strings and returning one string.
func str2(name string, fn func(a, b string) string) *starlark.Builtin {
	return builtin(name, func(_ *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var a, c string
		if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "a", &a, "b", &c); err != nil {
			return nil, err
		}
		return starlark.String(fn(a, c)), nil
	})
}

// bool2 builds a builtin taking two strings and returning a bool.
func bool2(name string, fn func(a, b string) bool) *starlark.Builtin {
	return builtin(name, func(_ *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var a, b string
		if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "str", &a, "substr", &b); err != nil {
			return nil, err
		}
		return starlark.Bool(fn(a, b)), nil
	})
}

// lenToResult builds a builtin taking a length and returning a (value, error)
// tuple, used by random.bytes and random.hex.
func lenToResult(name string, fn func(int) (string, error)) *starlark.Builtin {
	return builtin(name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var length int
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "length", &length); err != nil {
			return nil, err
		}
		s, err := fn(length)
		if err != nil {
			return errResult(err.Error()), nil
		}
		return okResult(starlark.String(s)), nil
	})
}

// iterableToStrings collects a Starlark iterable's elements as strings.
func iterableToStrings(v starlark.Value) []string {
	iterable, ok := v.(starlark.Iterable)
	if !ok {
		return nil
	}
	iter := iterable.Iterate()
	defer iter.Done()

	var out []string
	var elem starlark.Value
	for iter.Next(&elem) {
		out = append(out, asString(elem))
	}
	return out
}
