// Package store provides database interfaces and types for the FaaS application.
// It defines the DB interface for function, version, and execution operations,
// along with core data types used throughout the application.
package store

import (
	"context"
	"errors"
)

var (
	ErrFunctionNotFound  = errors.New("function not found")
	ErrVersionNotFound   = errors.New("version not found")
	ErrNoActiveVersion   = errors.New("no active version")
	ErrExecutionNotFound = errors.New("execution not found")
)

// DB defines the database interface for the FaaS API.
type DB interface {
	// CreateFunction creates a new function. Returns the created function with
	// timestamps populated.
	CreateFunction(ctx context.Context, fn Function) (Function, error)

	// GetFunction retrieves a function by ID.
	// Returns ErrFunctionNotFound if the function does not exist.
	GetFunction(ctx context.Context, id string) (Function, error)

	// ListFunctions returns paginated functions with their active versions.
	ListFunctions(ctx context.Context, params PaginationParams) ([]FunctionWithActiveVersion, int64, error)

	// UpdateFunction updates a function's fields.
	// Returns ErrFunctionNotFound if the function does not exist.
	UpdateFunction(ctx context.Context, id string, updates UpdateFunctionRequest) error

	// DeleteFunction removes a function and its associated data.
	// Returns ErrFunctionNotFound if the function does not exist.
	DeleteFunction(ctx context.Context, id string) error

	// CreateVersion creates a new version for a function and sets it as active.
	// Returns ErrFunctionNotFound if the function does not exist.
	CreateVersion(ctx context.Context, functionID string, code string, createdBy *string) (FunctionVersion, error)

	// GetVersion retrieves a specific version by function ID and version number.
	// Returns ErrVersionNotFound if the version does not exist.
	GetVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error)

	// GetVersionByID retrieves a version by its unique ID.
	// Returns ErrVersionNotFound if the version does not exist.
	GetVersionByID(ctx context.Context, versionID string) (FunctionVersion, error)

	// ListVersions returns paginated versions for a function.
	ListVersions(ctx context.Context, functionID string, params PaginationParams) ([]FunctionVersion, int64, error)

	// GetActiveVersion retrieves the currently active version for a function.
	// Returns ErrNoActiveVersion if no version is active.
	GetActiveVersion(ctx context.Context, functionID string) (FunctionVersion, error)

	// ActivateVersion sets a specific version as the active version.
	// Returns ErrVersionNotFound if the version does not exist.
	ActivateVersion(ctx context.Context, functionID string, version int) error

	// CreateExecution records a new execution. Returns the execution with
	// timestamps populated.
	CreateExecution(ctx context.Context, exec Execution) (Execution, error)

	// GetExecution retrieves an execution by ID.
	// Returns ErrExecutionNotFound if the execution does not exist.
	GetExecution(ctx context.Context, executionID string) (Execution, error)

	// UpdateExecution updates an execution's status and results.
	// Returns ErrExecutionNotFound if the execution does not exist.
	UpdateExecution(ctx context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string) error

	// ListExecutions returns paginated executions for a function.
	ListExecutions(ctx context.Context, functionID string, params PaginationParams) ([]Execution, int64, error)

	// DeleteOldExecutions removes executions older than the given timestamp.
	// Returns the number of deleted records.
	DeleteOldExecutions(ctx context.Context, beforeTimestamp int64) (int64, error)

	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error
}
