package api

import (
	"net/http"
)

// Server represents the API server
type Server struct {
	handler *Handler
	mux     *http.ServeMux
	db      DB
}

// NewServer creates a new API server with all routes configured
func NewServer(db DB) *Server {
	s := &Server{
		handler: NewHandler(db),
		mux:     http.NewServeMux(),
		db:      db,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Function Management
	s.mux.HandleFunc("POST /api/functions", s.handler.CreateFunction)
	s.mux.HandleFunc("GET /api/functions", s.handler.ListFunctions)
	s.mux.HandleFunc("GET /api/functions/{id}", s.handler.GetFunction)
	s.mux.HandleFunc("PUT /api/functions/{id}", s.handler.UpdateFunction)
	s.mux.HandleFunc("DELETE /api/functions/{id}", s.handler.DeleteFunction)
	s.mux.HandleFunc("PUT /api/functions/{id}/env", s.handler.UpdateEnvVars)

	// Version Management
	s.mux.HandleFunc("GET /api/functions/{id}/versions", s.handler.ListVersions)
	s.mux.HandleFunc("GET /api/functions/{id}/versions/{version}", s.handler.GetVersion)
	s.mux.HandleFunc("POST /api/functions/{id}/versions/{version}/activate", s.handler.ActivateVersion)
	s.mux.HandleFunc("GET /api/functions/{id}/diff/{v1}/{v2}", s.handler.GetVersionDiff)

	// Execution History
	s.mux.HandleFunc("GET /api/functions/{id}/executions", s.handler.ListExecutions)
	s.mux.HandleFunc("GET /api/executions/{id}", s.handler.GetExecution)
	s.mux.HandleFunc("GET /api/executions/{id}/logs", s.handler.GetExecutionLogs)

	// Runtime Execution (all HTTP methods)
	s.mux.HandleFunc("GET /fn/{function_id}", s.handler.ExecuteFunction)
	s.mux.HandleFunc("POST /fn/{function_id}", s.handler.ExecuteFunction)
	s.mux.HandleFunc("PUT /fn/{function_id}", s.handler.ExecuteFunction)
	s.mux.HandleFunc("DELETE /fn/{function_id}", s.handler.ExecuteFunction)
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
