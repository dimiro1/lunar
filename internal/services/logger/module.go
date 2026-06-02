package logger

import "go.uber.org/fx"

// Module provides the SQLite-backed execution logger, bound to the
// logger.Logger interface.
var Module = fx.Module("logger",
	fx.Provide(
		fx.Annotate(NewSQLiteLogger, fx.As(new(Logger))),
	),
)
