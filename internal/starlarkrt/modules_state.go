package starlarkrt

import (
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"go.starlark.net/starlark"
)

// logModule exposes structured logging: log.info/debug/warn/error(message).
func logModule(log logger.Logger, executionID string) starlark.Value {
	logFn := func(name string, emit func(executionID, message string)) *starlark.Builtin {
		return builtin("log."+name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var message string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "message", &message); err != nil {
				return nil, err
			}
			emit(executionID, message)
			return starlark.None, nil
		})
	}

	return module("log", starlark.StringDict{
		"info":  logFn("info", log.Info),
		"debug": logFn("debug", log.Debug),
		"warn":  logFn("warn", log.Warn),
		"error": logFn("error", log.Error),
	})
}

// kvModule exposes the per-function and global key-value store.
func kvModule(store kv.Store, functionID string) starlark.Value {
	getter := func(name string, lookup func(key string) (string, error)) *starlark.Builtin {
		return builtin("kv."+name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key); err != nil {
				return nil, err
			}
			value, err := lookup(key)
			if err != nil {
				return starlark.None, nil
			}
			return starlark.String(value), nil
		})
	}

	setter := func(name string, store func(key, value string) error) *starlark.Builtin {
		return builtin("kv."+name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key, value string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &value); err != nil {
				return nil, err
			}
			return starlark.Bool(store(key, value) == nil), nil
		})
	}

	deleter := func(name string, del func(key string) error) *starlark.Builtin {
		return builtin("kv."+name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key); err != nil {
				return nil, err
			}
			return starlark.Bool(del(key) == nil), nil
		})
	}

	lister := func(name string, list func() ([]string, error)) *starlark.Builtin {
		return builtin("kv."+name, func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			keys, err := list()
			if err != nil {
				return starlark.None, nil
			}
			elems := make([]starlark.Value, len(keys))
			for i, k := range keys {
				elems[i] = starlark.String(k)
			}
			return starlark.NewList(elems), nil
		})
	}

	aller := func(name string, all func() (map[string]string, error)) *starlark.Builtin {
		return builtin("kv."+name, func(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			m, err := all()
			if err != nil {
				return starlark.None, nil
			}
			return stringMapToDict(m), nil
		})
	}

	return module("kv", starlark.StringDict{
		"get":            getter("get", func(key string) (string, error) { return store.Get(functionID, key) }),
		"set":            setter("set", func(key, value string) error { return store.Set(functionID, key, value) }),
		"delete":         deleter("delete", func(key string) error { return store.Delete(functionID, key) }),
		"listKeys":       lister("listKeys", func() ([]string, error) { return store.ListKeys(functionID) }),
		"all":            aller("all", func() (map[string]string, error) { return store.All(functionID) }),
		"getGlobal":      getter("getGlobal", store.GetGlobal),
		"setGlobal":      setter("setGlobal", store.SetGlobal),
		"deleteGlobal":   deleter("deleteGlobal", store.DeleteGlobal),
		"listGlobalKeys": lister("listGlobalKeys", store.ListGlobalKeys),
		"allGlobal":      aller("allGlobal", store.AllGlobal),
	})
}

// envModule exposes per-function environment variables.
func envModule(store env.Store, functionID string) starlark.Value {
	return module("env", starlark.StringDict{
		"get": builtin("env.get", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key); err != nil {
				return nil, err
			}
			value, err := store.Get(functionID, key)
			if err != nil {
				return starlark.None, nil
			}
			return starlark.String(value), nil
		}),
		"set": builtin("env.set", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key, value string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &value); err != nil {
				return nil, err
			}
			return starlark.Bool(store.Set(functionID, key, value) == nil), nil
		}),
		"delete": builtin("env.delete", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var key string
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key); err != nil {
				return nil, err
			}
			return starlark.Bool(store.Delete(functionID, key) == nil), nil
		}),
	})
}
