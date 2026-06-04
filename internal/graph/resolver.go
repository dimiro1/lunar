package graph

import (
	"context"
	"errors"
	"fmt"

	internalcron "github.com/dimiro1/lunar/internal/cron"
	"github.com/dimiro1/lunar/internal/graph/model"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
)

//go:generate go tool gqlgen generate

// Resolver is the root resolver and the injection point for the dependencies the
// GraphQL resolvers need. gqlgen wires the generated Query/Mutation resolvers to
// this struct; fields added here are reachable from every resolver method.
//
// It is hand-written (gqlgen will not overwrite an existing resolver.go), so new
// dependencies are added by extending this struct and the fx provider in
// module.go.
type Resolver struct {
	DB           store.DB
	EnvStore     env.Store
	KVStore      kv.Store
	Scheduler    *internalcron.FunctionScheduler
	Logger       logger.Logger
	AITracker    ai.Tracker
	EmailTracker email.Tracker
}

// paginationParams builds normalized pagination parameters from optional
// GraphQL limit/offset arguments (which default via the schema but may be null).
func paginationParams(limit, offset *int) store.PaginationParams {
	p := store.PaginationParams{}
	if limit != nil {
		p.Limit = *limit
	}
	if offset != nil {
		p.Offset = *offset
	}
	return p.Normalize()
}

// pageInfo builds the PageInfo for a paginated response.
func pageInfo(total int64, p store.PaginationParams) *store.PaginationInfo {
	return &store.PaginationInfo{Total: total, Limit: p.Limit, Offset: p.Offset}
}

// loadFunction fetches a function together with its active version. It returns
// (nil, nil) when the function does not exist, which Query.function surfaces as
// a null result.
func (r *Resolver) loadFunction(ctx context.Context, id string) (*store.FunctionWithActiveVersion, error) {
	fn, err := r.DB.GetFunction(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrFunctionNotFound) {
			return nil, nil
		}
		return nil, err
	}
	active, err := r.DB.GetActiveVersion(ctx, id)
	if err != nil {
		return nil, err
	}
	return &store.FunctionWithActiveVersion{Function: fn, ActiveVersion: active}, nil
}

// reloadFunction is loadFunction for non-null results: mutations use it to return
// the affected function, treating a missing function as an error.
func (r *Resolver) reloadFunction(ctx context.Context, id string) (*store.FunctionWithActiveVersion, error) {
	fn, err := r.loadFunction(ctx, id)
	if err != nil {
		return nil, err
	}
	if fn == nil {
		return nil, fmt.Errorf("function %q not found", id)
	}
	return fn, nil
}

// cronStatusEnum converts the store's *string cron status into the CronStatus
// enum, mapping nil or empty (no schedule configured) to null.
func cronStatusEnum(s *string) *store.CronStatus {
	if s == nil || *s == "" {
		return nil
	}
	cs := store.CronStatus(*s)
	return &cs
}

// cronStatusString is the inverse of cronStatusEnum: it converts a CronStatus
// enum back to the *string the store layer expects.
func cronStatusString(c *store.CronStatus) *string {
	if c == nil {
		return nil
	}
	s := string(*c)
	return &s
}

// languageValue converts an optional Language enum to the string the store's
// CreateVersion expects, mapping nil to "" (which the store defaults to lua).
func languageValue(l *store.Language) string {
	if l == nil {
		return ""
	}
	return string(*l)
}

// loadExecution fetches an execution by ID, returning (nil, nil) when it does
// not exist so reverse edges (AIRequest.execution, …) resolve to null.
func (r *Resolver) loadExecution(ctx context.Context, id string) (*store.Execution, error) {
	exec, err := r.DB.GetExecution(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrExecutionNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &exec, nil
}

// loadVersionByID fetches a version by its ID, returning (nil, nil) when it does
// not exist so Execution.version resolves to null for a deleted version.
func (r *Resolver) loadVersionByID(ctx context.Context, versionID string) (*store.FunctionVersion, error) {
	v, err := r.DB.GetVersionByID(ctx, versionID)
	if err != nil {
		if errors.Is(err, store.ErrVersionNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// executionConnection lists a function's executions as a paginated connection.
// Shared by the top-level executions query and Function.executions.
func (r *Resolver) executionConnection(ctx context.Context, functionID string, limit, offset *int) (*model.ExecutionConnection, error) {
	params := paginationParams(limit, offset)
	executions, total, err := r.DB.ListExecutions(ctx, functionID, params)
	if err != nil {
		return nil, err
	}
	return &model.ExecutionConnection{Nodes: executions, PageInfo: pageInfo(total, params)}, nil
}

// logEntryConnection builds a paginated LogEntryConnection for an execution.
// Shared by the top-level executionLogs query and Execution.logs.
func (r *Resolver) logEntryConnection(executionID string, limit, offset *int) *model.LogEntryConnection {
	params := paginationParams(limit, offset)
	entries, total := r.Logger.EntriesPaginated(executionID, params.Limit, params.Offset)
	nodes := make([]model.LogEntry, len(entries))
	for i, e := range entries {
		nodes[i] = model.LogEntry{
			Level:     mapLogLevel(e.Level),
			Message:   e.Message,
			CreatedAt: int(e.Timestamp),
		}
	}
	return &model.LogEntryConnection{Nodes: nodes, PageInfo: pageInfo(total, params)}
}

// aiRequestConnection builds a paginated AIRequestConnection for an execution.
// Shared by the top-level executionAiRequests query and Execution.aiRequests.
// A nil AITracker (e.g. when AI tracking is not configured) yields an empty
// connection rather than panicking.
func (r *Resolver) aiRequestConnection(executionID string, limit, offset *int) *model.AIRequestConnection {
	params := paginationParams(limit, offset)
	if r.AITracker == nil {
		return &model.AIRequestConnection{PageInfo: pageInfo(0, params)}
	}
	requests, total := r.AITracker.RequestsPaginated(executionID, params.Limit, params.Offset)
	return &model.AIRequestConnection{Nodes: requests, PageInfo: pageInfo(total, params)}
}

// emailRequestConnection builds a paginated EmailRequestConnection for an
// execution. Shared by executionEmailRequests and Execution.emailRequests.
// A nil EmailTracker yields an empty connection rather than panicking.
func (r *Resolver) emailRequestConnection(executionID string, limit, offset *int) *model.EmailRequestConnection {
	params := paginationParams(limit, offset)
	if r.EmailTracker == nil {
		return &model.EmailRequestConnection{PageInfo: pageInfo(0, params)}
	}
	requests, total := r.EmailTracker.RequestsPaginated(executionID, params.Limit, params.Offset)
	return &model.EmailRequestConnection{Nodes: requests, PageInfo: pageInfo(total, params)}
}

// computeNextRun derives a function's next scheduled run from its cron settings.
// Shared by the top-level nextRun query and Function.nextRun.
func computeNextRun(fn store.Function) (*model.NextRun, error) {
	if fn.CronSchedule == nil || *fn.CronSchedule == "" {
		return &model.NextRun{HasSchedule: false}, nil
	}

	if fn.CronStatus == nil || *fn.CronStatus != string(store.CronStatusActive) {
		return &model.NextRun{
			HasSchedule:  true,
			CronSchedule: fn.CronSchedule,
			CronStatus:   cronStatusEnum(fn.CronStatus),
			IsPaused:     true,
		}, nil
	}

	nextRun, err := internalcron.GetNextRunFromSchedule(*fn.CronSchedule)
	if err != nil {
		return nil, err
	}

	out := &model.NextRun{
		HasSchedule:  true,
		CronSchedule: fn.CronSchedule,
		CronStatus:   cronStatusEnum(fn.CronStatus),
		IsPaused:     false,
	}
	if nextRun != nil {
		unix := int(nextRun.Unix())
		human := internalcron.FormatNextRun(*nextRun)
		out.NextRun = &unix
		out.NextRunHuman = &human
	}
	return out, nil
}

// mapLogLevel converts a logger severity to the GraphQL LogLevel enum.
func mapLogLevel(l logger.LogLevel) model.LogLevel {
	switch l {
	case logger.Debug:
		return model.LogLevelDebug
	case logger.Info:
		return model.LogLevelInfo
	case logger.Warn:
		return model.LogLevelWarn
	case logger.Error:
		return model.LogLevelError
	default:
		return model.LogLevelInfo
	}
}
