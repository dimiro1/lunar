package api

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	internalcron "github.com/dimiro1/lunar/internal/cron"
	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/graph"
	"github.com/dimiro1/lunar/internal/runner"
	"github.com/dimiro1/lunar/internal/starlarkrt"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
	"github.com/rs/xid"
)

// Server represents the API server. With the management API now served over
// GraphQL, the REST surface is limited to auth (login/device flow), the public
// /fn/* execution passthrough, and the frontend — so the server only holds the
// collaborators those routes and the GraphQL handler need.
type Server struct {
	mux             *http.ServeMux
	db              store.DB
	execDeps        *ExecuteFunctionDeps
	frontendHandler http.Handler
	apiKey          string
	httpServer      *http.Server
	deviceAuth      *DeviceAuthStore
	baseURL         string
	graphQL         *handler.Server
}

// ServerConfig holds configuration for creating a Server
type ServerConfig struct {
	DB               store.DB
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       internalhttp.Client
	AITracker        ai.Tracker
	EmailTracker     email.Tracker
	Scheduler        *internalcron.FunctionScheduler
	ExecutionTimeout time.Duration
	FrontendHandler  http.Handler
	APIKey           string
	BaseURL          string
}

// NewServer creates a new API server with full configuration. It constructs the
// AI/email clients, Lua runtime, and execution engine from the supplied
// ingredients, then assembles the server.
//
// The fx wiring in module.go builds those same collaborators as independent graph
// nodes and calls newServer directly, so this constructor remains the
// convenience entry point used by tests and any non-fx caller.
func NewServer(config ServerConfig) *Server {
	// Create AI and Email clients
	aiClient := ai.NewDefaultClient(config.HTTPClient, config.EnvStore)
	emailClient := email.NewDefaultClient(config.EnvStore)

	// Create the language runtimes (Lua and Starlark share the same collaborators).
	luaRuntime := runner.NewLuaRuntime(runner.LuaRuntimeConfig{
		Logger:       config.Logger,
		KV:           config.KVStore,
		Env:          config.EnvStore,
		HTTP:         config.HTTPClient,
		AI:           aiClient,
		AITracker:    config.AITracker,
		Email:        emailClient,
		EmailTracker: config.EmailTracker,
		Timeout:      config.ExecutionTimeout,
	})
	starlarkRuntime := starlarkrt.New(starlarkrt.Config{
		Logger:       config.Logger,
		KV:           config.KVStore,
		Env:          config.EnvStore,
		HTTP:         config.HTTPClient,
		AI:           aiClient,
		AITracker:    config.AITracker,
		Email:        emailClient,
		EmailTracker: config.EmailTracker,
		Timeout:      config.ExecutionTimeout,
	})

	// Create execution engine
	eng := engine.New(engine.Config{
		DB: config.DB,
		Runtimes: []engine.RuntimeEntry{
			{Language: engine.LanguageLua, Runtime: luaRuntime},
			{Language: engine.LanguageStarlark, Runtime: starlarkRuntime},
		},
		Logger:           config.Logger,
		KVStore:          config.KVStore,
		EnvStore:         config.EnvStore,
		HTTPClient:       config.HTTPClient,
		AIClient:         aiClient,
		AITracker:        config.AITracker,
		EmailClient:      emailClient,
		EmailTracker:     config.EmailTracker,
		ExecutionTimeout: config.ExecutionTimeout,
		IDGenerator:      func() string { return xid.New().String() },
	})

	return newServer(serverDeps{
		DB:              config.DB,
		Engine:          eng,
		FrontendHandler: config.FrontendHandler,
		APIKey:          config.APIKey,
		BaseURL:         config.BaseURL,
		GraphQL: graph.NewServer(&graph.Resolver{
			DB:           config.DB,
			EnvStore:     config.EnvStore,
			KVStore:      config.KVStore,
			Scheduler:    config.Scheduler,
			Logger:       config.Logger,
			AITracker:    config.AITracker,
			EmailTracker: config.EmailTracker,
		}),
	})
}

// serverDeps are the fully-constructed collaborators a Server needs. Unlike
// ServerConfig — which carries the raw ingredients used to build the engine and
// the GraphQL resolver — serverDeps takes the engine.Engine and GraphQL handler
// already assembled. This is the seam the fx graph injects through.
type serverDeps struct {
	DB              store.DB
	Engine          engine.Engine
	FrontendHandler http.Handler
	APIKey          string
	BaseURL         string
	GraphQL         *handler.Server
}

// newServer assembles a Server from its constructed dependencies and registers
// the routes.
func newServer(d serverDeps) *Server {
	s := &Server{
		mux:             http.NewServeMux(),
		db:              d.DB,
		execDeps:        &ExecuteFunctionDeps{Engine: d.Engine, BaseURL: d.BaseURL},
		frontendHandler: d.FrontendHandler,
		apiKey:          d.APIKey,
		deviceAuth:      NewDeviceAuthStore(),
		baseURL:         d.BaseURL,
		graphQL:         d.GraphQL,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes using functional handlers
func (s *Server) setupRoutes() {
	// Auth routes (no authentication required)
	s.mux.HandleFunc("POST /api/auth/login", HandleLogin(s.apiKey))
	s.mux.HandleFunc("POST /api/auth/logout", HandleLogout())

	// Device auth flow (no auth required for request and poll)
	s.mux.HandleFunc("POST /api/auth/device-request", HandleDeviceRequest(s.deviceAuth, s.baseURL))
	s.mux.HandleFunc("GET /api/auth/device-token", HandleDeviceToken(s.deviceAuth))

	// Protected routes - wrap with auth middleware
	authMiddleware := AuthMiddleware(s.apiKey, s.db)

	// Device approval (auth required - user must be logged in via the SPA)
	s.mux.Handle("GET /api/auth/device-approve", authMiddleware(http.HandlerFunc(HandleDeviceApproveStatus(s.deviceAuth))))
	s.mux.Handle("POST /api/auth/device-approve", authMiddleware(http.HandlerFunc(HandleDeviceApprove(s.deviceAuth, s.db))))

	// GraphQL API — the entire management surface (functions, versions,
	// executions, tokens). Query execution (POST) is auth-protected; the
	// GraphiQL playground UI (GET) is served publicly and posts back to
	// /graphql, which still enforces authentication.
	if s.graphQL != nil {
		s.mux.Handle("POST /graphql", authMiddleware(s.graphQL))
		s.mux.HandleFunc("GET /graphql", playground.Handler("Lunar GraphQL", "/graphql"))
	}

	// Runtime Execution - needs all dependencies (NO AUTH - public endpoint)
	// Register both exact match and wildcard patterns for routing support
	executeHandler := ExecuteFunctionHandler(*s.execDeps)
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		s.mux.HandleFunc(method+" /fn/{function_id}", executeHandler)
		s.mux.HandleFunc(method+" /fn/{function_id}/{path...}", executeHandler)
	}

	// Serve frontend files (catch-all route for SPA)
	if s.frontendHandler != nil {
		s.mux.Handle("/", s.frontendHandler)
	}
}

// Handler returns the http.Handler with all middleware applied
func (s *Server) Handler() http.Handler {
	return Chain(
		s.mux,
		RecoveryMiddleware,
		LoggingMiddleware,
		CORSMiddleware,
	)
}

// ListenAndServe starts the HTTP server on the specified address
func (s *Server) ListenAndServe(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting active connections
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
