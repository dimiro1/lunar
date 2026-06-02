package ai

import "go.uber.org/fx"

// Module provides the AI client and request tracker, bound to the ai.Client and
// ai.Tracker interfaces. The client depends on an http.Client and env.Store; the
// tracker depends on a *sql.DB — all supplied by other modules.
var Module = fx.Module("ai",
	fx.Provide(
		fx.Annotate(NewDefaultClient, fx.As(new(Client))),
		fx.Annotate(NewSQLiteTracker, fx.As(new(Tracker))),
	),
)
