package env

import "go.uber.org/fx"

// Module provides the SQLite-backed environment-variable store, bound to the
// env.Store interface.
var Module = fx.Module("env",
	fx.Provide(
		fx.Annotate(NewSQLiteStore, fx.As(new(Store))),
	),
)
