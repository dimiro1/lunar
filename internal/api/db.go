package api

import "context"

// DB defines the database interface for the FaaS API
type DB interface {
	// Function operations
	CreateFunction(ctx context.Context, fn Function) (Function, error)
	GetFunction(ctx context.Context, id string) (Function, error)
	ListFunctions(ctx context.Context, params PaginationParams) ([]Function, int64, error)
	UpdateFunction(ctx context.Context, id string, updates UpdateFunctionRequest) error
	DeleteFunction(ctx context.Context, id string) error
	UpdateFunctionEnvVars(ctx context.Context, id string, envVars map[string]string) error

	// Version operations
	CreateVersion(ctx context.Context, functionID string, code string, createdBy *string) (FunctionVersion, error)
	GetVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error)
	GetVersionByID(ctx context.Context, versionID string) (FunctionVersion, error)
	ListVersions(ctx context.Context, functionID string, params PaginationParams) ([]FunctionVersion, int64, error)
	GetActiveVersion(ctx context.Context, functionID string) (FunctionVersion, error)
	ActivateVersion(ctx context.Context, functionID string, version int) error

	// Execution operations
	CreateExecution(ctx context.Context, exec Execution) (Execution, error)
	GetExecution(ctx context.Context, executionID string) (Execution, error)
	UpdateExecution(ctx context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string) error
	ListExecutions(ctx context.Context, functionID string, params PaginationParams) ([]Execution, int64, error)

	// Health check
	Ping(ctx context.Context) error
}
