package http

import "go.uber.org/fx"

// Module provides the default outbound HTTP client, bound to the http.Client
// interface.
var Module = fx.Module("http",
	fx.Provide(
		fx.Annotate(NewDefaultClient, fx.As(new(Client))),
	),
)
