package engine

import (
	"github.com/dimiro1/lunar/internal/config"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
	"github.com/rs/xid"
	"go.uber.org/fx"
)

// Module provides the execution engine, bound to the engine.Engine interface.
var Module = fx.Module("engine",
	fx.Provide(provideEngine),
)

// engineParams gathers the engine's collaborators plus the app config (for the
// execution timeout) via dependency injection.
type engineParams struct {
	fx.In

	DB           store.DB
	Runtime      Runtime
	Logger       logger.Logger
	KVStore      kv.Store
	EnvStore     env.Store
	HTTPClient   http.Client
	AIClient     ai.Client
	AITracker    ai.Tracker
	EmailClient  email.Client
	EmailTracker email.Tracker
	Config       config.Config
}

func provideEngine(p engineParams) Engine {
	return New(Config{
		DB:               p.DB,
		Runtime:          p.Runtime,
		Logger:           p.Logger,
		KVStore:          p.KVStore,
		EnvStore:         p.EnvStore,
		HTTPClient:       p.HTTPClient,
		AIClient:         p.AIClient,
		AITracker:        p.AITracker,
		EmailClient:      p.EmailClient,
		EmailTracker:     p.EmailTracker,
		ExecutionTimeout: p.Config.ExecutionTimeout,
		IDGenerator:      func() string { return xid.New().String() },
	})
}
