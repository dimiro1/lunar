package api

import (
	"net/http"
	"time"

	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
)

// Server represents the API server
type Server struct {
	mux             *http.ServeMux
	db              DB
	execDeps        *ExecuteFunctionDeps
	logger          logger.Logger
	frontendHandler http.Handler
}

// ServerConfig holds configuration for creating a Server
type ServerConfig struct {
	DB               DB
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       internalhttp.Client
	ExecutionTimeout time.Duration
	FrontendHandler  http.Handler
}

// NewServer creates a new API server with full configuration
func NewServer(config ServerConfig) *Server {
	execDeps := &ExecuteFunctionDeps{
		DB:               config.DB,
		Logger:           config.Logger,
		KVStore:          config.KVStore,
		EnvStore:         config.EnvStore,
		HTTPClient:       config.HTTPClient,
		ExecutionTimeout: config.ExecutionTimeout,
	}

	s := &Server{
		mux:             http.NewServeMux(),
		db:              config.DB,
		execDeps:        execDeps,
		logger:          config.Logger,
		frontendHandler: config.FrontendHandler,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes using functional handlers
func (s *Server) setupRoutes() {
	// Function Management - only need DB
	s.mux.HandleFunc("POST /api/functions", CreateFunctionHandler(s.db))
	s.mux.HandleFunc("GET /api/functions", ListFunctionsHandler(s.db))
	s.mux.HandleFunc("GET /api/functions/{id}", GetFunctionHandler(s.db))
	s.mux.HandleFunc("PUT /api/functions/{id}", UpdateFunctionHandler(s.db))
	s.mux.HandleFunc("DELETE /api/functions/{id}", DeleteFunctionHandler(s.db))
	s.mux.HandleFunc("PUT /api/functions/{id}/env", UpdateEnvVarsHandler(s.db))

	// Version Management - only need DB
	s.mux.HandleFunc("GET /api/functions/{id}/versions", ListVersionsHandler(s.db))
	s.mux.HandleFunc("GET /api/functions/{id}/versions/{version}", GetVersionHandler(s.db))
	s.mux.HandleFunc("POST /api/functions/{id}/versions/{version}/activate", ActivateVersionHandler(s.db))
	s.mux.HandleFunc("GET /api/functions/{id}/diff/{v1}/{v2}", GetVersionDiffHandler(s.db))

	// Execution History - only need DB
	s.mux.HandleFunc("GET /api/functions/{id}/executions", ListExecutionsHandler(s.db))
	s.mux.HandleFunc("GET /api/executions/{id}", GetExecutionHandler(s.db))
	s.mux.HandleFunc("GET /api/executions/{id}/logs", GetExecutionLogsHandler(s.db, s.logger))

	// Runtime Execution - needs all dependencies
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
	return http.ListenAndServe(addr, s.Handler())
}
