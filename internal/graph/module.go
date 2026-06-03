package graph

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	internalcron "github.com/dimiro1/lunar/internal/cron"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/fx"
)

// Module provides the GraphQL resolver root and the gqlgen HTTP handler. The
// handler is consumed by the api module, which mounts it at /graphql; this
// module deliberately owns no lifecycle of its own.
var Module = fx.Module("graphql",
	fx.Provide(
		newResolver,
		NewServer,
	),
)

// resolverParams gathers the resolver's dependencies from the fx graph.
type resolverParams struct {
	fx.In

	DB           store.DB
	EnvStore     env.Store
	KVStore      kv.Store
	Scheduler    *internalcron.FunctionScheduler
	Logger       logger.Logger
	AITracker    ai.Tracker
	EmailTracker email.Tracker
}

// newResolver builds the root resolver from its injected dependencies. New
// resolver dependencies are added as fields on Resolver and on resolverParams.
func newResolver(p resolverParams) *Resolver {
	return &Resolver{
		DB:           p.DB,
		EnvStore:     p.EnvStore,
		KVStore:      p.KVStore,
		Scheduler:    p.Scheduler,
		Logger:       p.Logger,
		AITracker:    p.AITracker,
		EmailTracker: p.EmailTracker,
	}
}

// NewServer assembles the gqlgen HTTP handler from the executable schema. It
// enables the GET and POST transports, an LRU query cache, and schema
// introspection (used by the playground and by generated clients).
func NewServer(r *Resolver) *handler.Server {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: r}))

	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})

	return srv
}
