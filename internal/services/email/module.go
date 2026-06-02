package email

import "go.uber.org/fx"

// Module provides the email client and request tracker, bound to the
// email.Client and email.Tracker interfaces.
var Module = fx.Module("email",
	fx.Provide(
		fx.Annotate(NewDefaultClient, fx.As(new(Client))),
		fx.Annotate(NewSQLiteTracker, fx.As(new(Tracker))),
	),
)
