package cron

import (
	"context"

	"github.com/dimiro1/lunar/internal/config"
	"github.com/dimiro1/lunar/internal/store"
	"go.uber.org/fx"
)

// Module provides the function cron scheduler and ties its start/stop to the
// application lifecycle.
var Module = fx.Module("cron",
	fx.Provide(provideScheduler),
	fx.Invoke(registerScheduler),
)

func provideScheduler(db store.DB, cfg config.Config) *FunctionScheduler {
	return NewScheduler(db, cfg.BaseURL)
}

func registerScheduler(lc fx.Lifecycle, s *FunctionScheduler) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { return s.Start() },
		OnStop: func(context.Context) error {
			s.Stop()
			return nil
		},
	})
}
