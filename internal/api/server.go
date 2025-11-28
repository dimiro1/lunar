package api

import (
	"context"
	"net/http"
	"time"

	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	"github.com/dimiro1/faas-go/internal/store"
)

// Server represents the API server
type Server struct {
	mux             *http.ServeMux
	db              store.DB
	execDeps        *ExecuteFunctionDeps
	logger          logger.Logger
	aiTracker       ai.Tracker
	frontendHandler http.Handler
	apiKey          string
	httpServer      *http.Server
}

// ServerConfig holds configuration for creating a Server
type ServerConfig struct {
	DB               store.DB
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       internalhttp.Client
	AITracker        ai.Tracker
	ExecutionTimeout time.Duration
	FrontendHandler  http.Handler
	APIKey           string
	BaseURL          string
}

// NewServer creates a new API server with full configuration
func NewServer(config ServerConfig) *Server {
	execDeps := &ExecuteFunctionDeps{
		DB:               config.DB,
		Logger:           config.Logger,
		KVStore:          config.KVStore,
		EnvStore:         config.EnvStore,
		HTTPClient:       config.HTTPClient,
		AIClient:         ai.NewDefaultClient(config.HTTPClient, config.EnvStore),
		AITracker:        config.AITracker,
		ExecutionTimeout: config.ExecutionTimeout,
		BaseURL:          config.BaseURL,
	}

	s := &Server{
		mux:             http.NewServeMux(),
		db:              config.DB,
		execDeps:        execDeps,
		logger:          config.Logger,
		aiTracker:       config.AITracker,
		frontendHandler: config.FrontendHandler,
		apiKey:          config.APIKey,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes using functional handlers
func (s *Server) setupRoutes() {
	// Auth routes (no authentication required)
	s.mux.HandleFunc("POST /api/auth/login", HandleLogin(s.apiKey))
	s.mux.HandleFunc("POST /api/auth/logout", HandleLogout())

	// API documentation (no authentication required)
	s.mux.HandleFunc("GET /docs", docsPageHandler)
	s.mux.HandleFunc("HEAD /docs", docsPageHandler)
	s.mux.HandleFunc("GET /docs/openapi.yaml", openAPISpecHandler)
	s.mux.HandleFunc("HEAD /docs/openapi.yaml", openAPISpecHandler)

	// Protected API routes - wrap with auth middleware
	authMiddleware := AuthMiddleware(s.apiKey)

	// Function Management - only need DB
	s.mux.Handle("POST /api/functions", authMiddleware(http.HandlerFunc(CreateFunctionHandler(s.db))))
	s.mux.Handle("GET /api/functions", authMiddleware(http.HandlerFunc(ListFunctionsHandler(s.db))))
	s.mux.Handle("GET /api/functions/{id}", authMiddleware(http.HandlerFunc(GetFunctionHandler(s.db, s.execDeps.EnvStore))))
	s.mux.Handle("PUT /api/functions/{id}", authMiddleware(http.HandlerFunc(UpdateFunctionHandler(s.db))))
	s.mux.Handle("DELETE /api/functions/{id}", authMiddleware(http.HandlerFunc(DeleteFunctionHandler(s.db))))
	s.mux.Handle("PUT /api/functions/{id}/env", authMiddleware(http.HandlerFunc(UpdateEnvVarsHandler(s.db, s.execDeps.EnvStore))))

	// Version Management - only need DB
	s.mux.Handle("GET /api/functions/{id}/versions", authMiddleware(http.HandlerFunc(ListVersionsHandler(s.db))))
	s.mux.Handle("GET /api/functions/{id}/versions/{version}", authMiddleware(http.HandlerFunc(GetVersionHandler(s.db))))
	s.mux.Handle("POST /api/functions/{id}/versions/{version}/activate", authMiddleware(http.HandlerFunc(ActivateVersionHandler(s.db))))
	s.mux.Handle("GET /api/functions/{id}/diff/{v1}/{v2}", authMiddleware(http.HandlerFunc(GetVersionDiffHandler(s.db))))

	// Execution History - only need DB
	s.mux.Handle("GET /api/functions/{id}/executions", authMiddleware(http.HandlerFunc(ListExecutionsHandler(s.db))))
	s.mux.Handle("GET /api/executions/{id}", authMiddleware(http.HandlerFunc(GetExecutionHandler(s.db))))
	s.mux.Handle("GET /api/executions/{id}/logs", authMiddleware(http.HandlerFunc(GetExecutionLogsHandler(s.db, s.logger))))
	s.mux.Handle("GET /api/executions/{id}/ai-requests", authMiddleware(http.HandlerFunc(GetExecutionAIRequestsHandler(s.db, s.aiTracker))))

	// Runtime Execution - needs all dependencies (NO AUTH - public endpoint)
	executeHandler := ExecuteFunctionHandler(*s.execDeps)
	s.mux.HandleFunc("GET /fn/{function_id}", executeHandler)
	s.mux.HandleFunc("POST /fn/{function_id}", executeHandler)
	s.mux.HandleFunc("PUT /fn/{function_id}", executeHandler)
	s.mux.HandleFunc("DELETE /fn/{function_id}", executeHandler)

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
