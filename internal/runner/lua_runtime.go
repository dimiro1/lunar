package runner

import (
	"context"
	"time"

	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
)

// Compile-time check that LuaRuntime implements engine.Runtime
var _ engine.Runtime = (*LuaRuntime)(nil)

// LuaRuntime implements the engine.Runtime interface for Lua code execution.
type LuaRuntime struct {
	logger       logger.Logger
	kv           kv.Store
	env          env.Store
	http         internalhttp.Client
	ai           ai.Client
	aiTracker    ai.Tracker
	email        email.Client
	emailTracker email.Tracker
	timeout      time.Duration
}

// LuaRuntimeConfig holds the configuration for creating a LuaRuntime.
type LuaRuntimeConfig struct {
	Logger       logger.Logger
	KV           kv.Store
	Env          env.Store
	HTTP         internalhttp.Client
	AI           ai.Client
	AITracker    ai.Tracker
	Email        email.Client
	EmailTracker email.Tracker
	Timeout      time.Duration
}

// NewLuaRuntime creates a new LuaRuntime with the given configuration.
func NewLuaRuntime(cfg LuaRuntimeConfig) *LuaRuntime {
	return &LuaRuntime{
		logger:       cfg.Logger,
		kv:           cfg.KV,
		env:          cfg.Env,
		http:         cfg.HTTP,
		ai:           cfg.AI,
		aiTracker:    cfg.AITracker,
		email:        cfg.Email,
		emailTracker: cfg.EmailTracker,
		timeout:      cfg.Timeout,
	}
}

// Execute implements the engine.Runtime interface.
func (r *LuaRuntime) Execute(ctx context.Context, req engine.RuntimeRequest) (*engine.RuntimeResult, error) {
	deps := Dependencies{
		Logger:       r.logger,
		KV:           r.kv,
		Env:          r.env,
		HTTP:         r.http,
		AI:           r.ai,
		AITracker:    r.aiTracker,
		Email:        r.email,
		EmailTracker: r.emailTracker,
		Timeout:      r.timeout,
	}

	runReq := Request{
		Context: req.Context,
		Event:   req.Event,
		Code:    req.Code,
	}

	resp, err := Run(ctx, deps, runReq)
	if err != nil {
		return nil, err
	}

	return &engine.RuntimeResult{
		Response: resp.HTTP,
	}, nil
}
