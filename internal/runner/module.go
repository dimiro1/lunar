package runner

import (
	"github.com/dimiro1/lunar/internal/config"
	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"go.uber.org/fx"
)

// Module provides the Lua runtime, bound to the engine.Runtime interface the
// engine consumes.
var Module = fx.Module("runner",
	fx.Provide(provideRuntime),
)

// runtimeParams gathers the runtime's collaborators plus the app config (for the
// execution timeout) via dependency injection.
type runtimeParams struct {
	fx.In

	Logger       logger.Logger
	KV           kv.Store
	Env          env.Store
	HTTP         internalhttp.Client
	AI           ai.Client
	AITracker    ai.Tracker
	Email        email.Client
	EmailTracker email.Tracker
	Config       config.Config
}

func provideRuntime(p runtimeParams) engine.Runtime {
	return NewLuaRuntime(LuaRuntimeConfig{
		Logger:       p.Logger,
		KV:           p.KV,
		Env:          p.Env,
		HTTP:         p.HTTP,
		AI:           p.AI,
		AITracker:    p.AITracker,
		Email:        p.Email,
		EmailTracker: p.EmailTracker,
		Timeout:      p.Config.ExecutionTimeout,
	})
}
