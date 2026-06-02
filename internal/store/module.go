package store

import "go.uber.org/fx"

// Module provides the SQLite-backed store, bound to the store.DB interface.
// It depends on a *sql.DB supplied by the composition root.
var Module = fx.Module("store",
	fx.Provide(
		fx.Annotate(NewSQLiteDB, fx.As(new(DB))),
	),
)
