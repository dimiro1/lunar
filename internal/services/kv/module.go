package kv

import "go.uber.org/fx"

// Module provides the SQLite-backed key/value store, bound to the kv.Store
// interface.
var Module = fx.Module("kv",
	fx.Provide(
		fx.Annotate(NewSQLiteStore, fx.As(new(Store))),
	),
)
