package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryDB is an in-memory implementation of the DB interface
type MemoryDB struct {
	mu         sync.RWMutex
	functions  map[string]Function
	versions   map[string][]FunctionVersion // functionID -> versions
	executions map[string]Execution         // executionID -> execution
	logs       map[string][]LogEntry        // executionID -> logs
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		functions:  make(map[string]Function),
		versions:   make(map[string][]FunctionVersion),
		executions: make(map[string]Execution),
		logs:       make(map[string][]LogEntry),
	}
}

// Function operations

func (db *MemoryDB) CreateFunction(ctx context.Context, fn Function) (Function, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	fn.CreatedAt = time.Now().Unix()
	fn.UpdatedAt = fn.CreatedAt
	if fn.EnvVars == nil {
		fn.EnvVars = make(map[string]string)
	}

	db.functions[fn.ID] = fn
	return fn, nil
}

func (db *MemoryDB) GetFunction(ctx context.Context, id string) (Function, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	fn, ok := db.functions[id]
	if !ok {
		return Function{}, fmt.Errorf("function not found")
	}
	return fn, nil
}

func (db *MemoryDB) ListFunctions(ctx context.Context) ([]Function, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	functions := make([]Function, 0, len(db.functions))
	for _, fn := range db.functions {
		functions = append(functions, fn)
	}
	return functions, nil
}

func (db *MemoryDB) UpdateFunction(ctx context.Context, id string, updates UpdateFunctionRequest) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	fn, ok := db.functions[id]
	if !ok {
		return fmt.Errorf("function not found")
	}

	if updates.Name != nil {
		fn.Name = *updates.Name
	}
	if updates.Description != nil {
		fn.Description = updates.Description
	}

	fn.UpdatedAt = time.Now().Unix()
	db.functions[id] = fn
	return nil
}

func (db *MemoryDB) DeleteFunction(ctx context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.functions[id]; !ok {
		return fmt.Errorf("function not found")
	}

	delete(db.functions, id)
	delete(db.versions, id)
	return nil
}

func (db *MemoryDB) UpdateFunctionEnvVars(ctx context.Context, id string, envVars map[string]string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	fn, ok := db.functions[id]
	if !ok {
		return fmt.Errorf("function not found")
	}

	fn.EnvVars = envVars
	fn.UpdatedAt = time.Now().Unix()
	db.functions[id] = fn
	return nil
}

// Version operations

func (db *MemoryDB) CreateVersion(ctx context.Context, functionID string, code string, createdBy *string) (FunctionVersion, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.functions[functionID]; !ok {
		return FunctionVersion{}, fmt.Errorf("function not found")
	}

	versions := db.versions[functionID]
	versionNum := len(versions) + 1

	// Deactivate all previous versions
	for i := range versions {
		versions[i].IsActive = false
	}

	version := FunctionVersion{
		ID:         fmt.Sprintf("ver_%s_v%d", functionID, versionNum),
		FunctionID: functionID,
		Version:    versionNum,
		Code:       code,
		CreatedAt:  time.Now().Unix(),
		CreatedBy:  createdBy,
		IsActive:   true,
	}

	versions = append(versions, version)
	db.versions[functionID] = versions

	return version, nil
}

func (db *MemoryDB) GetVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	versions := db.versions[functionID]
	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return FunctionVersion{}, fmt.Errorf("version not found")
}

func (db *MemoryDB) GetVersionByID(ctx context.Context, versionID string) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, versions := range db.versions {
		for _, v := range versions {
			if v.ID == versionID {
				return v, nil
			}
		}
	}

	return FunctionVersion{}, fmt.Errorf("version not found")
}

func (db *MemoryDB) ListVersions(ctx context.Context, functionID string) ([]FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	versions := db.versions[functionID]
	if versions == nil {
		return []FunctionVersion{}, nil
	}

	// Return in reverse order (newest first)
	result := make([]FunctionVersion, len(versions))
	for i, v := range versions {
		result[len(versions)-1-i] = v
	}

	return result, nil
}

func (db *MemoryDB) GetActiveVersion(ctx context.Context, functionID string) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	versions := db.versions[functionID]
	for _, v := range versions {
		if v.IsActive {
			return v, nil
		}
	}

	return FunctionVersion{}, fmt.Errorf("no active version found")
}

func (db *MemoryDB) ActivateVersion(ctx context.Context, functionID string, version int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	versions := db.versions[functionID]
	found := false

	for i := range versions {
		if versions[i].Version == version {
			versions[i].IsActive = true
			found = true
		} else {
			versions[i].IsActive = false
		}
	}

	if !found {
		return fmt.Errorf("version not found")
	}

	db.versions[functionID] = versions
	return nil
}

// Execution operations

func (db *MemoryDB) CreateExecution(ctx context.Context, exec Execution) (Execution, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec.CreatedAt = time.Now().Unix()
	db.executions[exec.ExecutionID] = exec
	return exec, nil
}

func (db *MemoryDB) GetExecution(ctx context.Context, executionID string) (Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, ok := db.executions[executionID]
	if !ok {
		return Execution{}, fmt.Errorf("execution not found")
	}
	return exec, nil
}

func (db *MemoryDB) UpdateExecution(ctx context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, ok := db.executions[executionID]
	if !ok {
		return fmt.Errorf("execution not found")
	}

	exec.Status = status
	exec.DurationMs = durationMs
	exec.ErrorMessage = errorMsg
	db.executions[executionID] = exec

	return nil
}

func (db *MemoryDB) ListExecutions(ctx context.Context, functionID string, limit int) ([]ExecutionWithLogCount, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var executions []ExecutionWithLogCount
	for _, exec := range db.executions {
		if exec.FunctionID == functionID {
			logCount := int64(len(db.logs[exec.ExecutionID]))
			executions = append(executions, ExecutionWithLogCount{
				Execution: exec,
				LogCount:  logCount,
			})
		}
	}

	// Apply limit
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

// Log operations

func (db *MemoryDB) CreateLog(ctx context.Context, log LogEntry) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	log.Timestamp = time.Now().Unix()
	db.logs[log.ExecutionID] = append(db.logs[log.ExecutionID], log)
	return nil
}

func (db *MemoryDB) GetExecutionLogs(ctx context.Context, executionID string) ([]LogEntry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	logs := db.logs[executionID]
	if logs == nil {
		return []LogEntry{}, nil
	}
	return logs, nil
}

func (db *MemoryDB) GetLogCount(ctx context.Context, executionID string) (int64, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return int64(len(db.logs[executionID])), nil
}

// Health check

func (db *MemoryDB) Ping(ctx context.Context) error {
	return nil
}
