package housekeeping

import (
	"context"

	"github.com/dimiro1/lunar/internal/store"
	"go.uber.org/fx"
)

// Module provides the housekeeping scheduler and ties its start/stop to the
// application lifecycle. The fx.Invoke forces construction because nothing else
// in the graph depends on the scheduler.
var Module = fx.Module("housekeeping",
	fx.Provide(provideScheduler),
	fx.Invoke(registerScheduler),
)

func provideScheduler(db store.DB) *Scheduler {
	return NewScheduler(db)
}

func registerScheduler(lc fx.Lifecycle, s *Scheduler) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { return s.Start() },
		OnStop: func(context.Context) error {
			s.Stop()
			return nil
		},
	})
}
