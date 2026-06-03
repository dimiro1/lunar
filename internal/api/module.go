package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/dimiro1/lunar/internal/config"
	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/store"
	"go.uber.org/fx"
)

// Module provides the HTTP API server and ties its start/stop to the
// application lifecycle.
var Module = fx.Module("api",
	fx.Provide(provideServer),
	fx.Invoke(registerServer),
)

// serverParams gathers everything the API server needs via dependency
// injection. The engine.Engine and the GraphQL handler are injected as
// fully-built graph nodes rather than assembled inside the server.
type serverParams struct {
	fx.In

	DB       store.DB
	Engine   engine.Engine
	Frontend http.Handler
	GraphQL  *handler.Server
	Config   config.Config
}

func provideServer(p serverParams) *Server {
	return newServer(serverDeps{
		DB:              p.DB,
		Engine:          p.Engine,
		FrontendHandler: p.Frontend,
		APIKey:          p.Config.APIKey,
		BaseURL:         p.Config.BaseURL,
		GraphQL:         p.GraphQL,
	})
}

// registerServer starts the HTTP server on app start and shuts it down
// gracefully on app stop. ListenAndServe blocks, so it runs in a goroutine; a
// bind failure triggers an app-wide shutdown with a non-zero exit code.
func registerServer(lc fx.Lifecycle, s *Server, cfg config.Config, shutdowner fx.Shutdowner) {
	addr := ":" + cfg.Port
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			slog.Info("Starting Lunar server",
				"port", cfg.Port,
				"data_dir", cfg.DataDir,
				"execution_timeout", cfg.ExecutionTimeout)
			slog.Info("Frontend available", "url", "http://localhost:"+cfg.Port)
			slog.Info("API available", "url", "http://localhost:"+cfg.Port+"/api")

			go func() {
				if err := s.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
					slog.Error("Server failed", "error", err)
					_ = shutdowner.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down server gracefully...")
			return s.Shutdown(ctx)
		},
	})
}
