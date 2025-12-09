package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var _ DB = (*MemoryDB)(nil)

// MemoryDB is an in-memory implementation of the DB interface
type MemoryDB struct {
	mu         sync.RWMutex
	functions  map[string]Function
	versions   map[string][]FunctionVersion // functionID -> versions
	executions map[string]Execution         // id -> execution
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		functions:  make(map[string]Function),
		versions:   make(map[string][]FunctionVersion),
		executions: make(map[string]Execution),
	}
}

// Function operations

func (db *MemoryDB) CreateFunction(_ context.Context, fn Function) (Function, error) {
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

func (db *MemoryDB) GetFunction(_ context.Context, id string) (Function, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	fn, ok := db.functions[id]
	if !ok {
		return Function{}, ErrFunctionNotFound
	}
	return fn, nil
}

func (db *MemoryDB) ListFunctions(_ context.Context, params PaginationParams) ([]FunctionWithActiveVersion, int64, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Normalize pagination parameters
	params = params.Normalize()

	// Get all functions with their active versions
	allFunctions := make([]FunctionWithActiveVersion, 0, len(db.functions))
	for _, fn := range db.functions {
		fnWithVersion := FunctionWithActiveVersion{
			Function: fn,
		}

		// Find active version
		if versions, ok := db.versions[fn.ID]; ok {
			for _, v := range versions {
				if v.IsActive {
					fnWithVersion.ActiveVersion = v
					break
				}
			}
		}

		allFunctions = append(allFunctions, fnWithVersion)
	}

	total := int64(len(allFunctions))

	// Apply pagination
	start := params.Offset
	if start > len(allFunctions) {
		return []FunctionWithActiveVersion{}, total, nil
	}

	end := start + params.Limit
	if end > len(allFunctions) {
		end = len(allFunctions)
	}

	return allFunctions[start:end], total, nil
}

func (db *MemoryDB) UpdateFunction(_ context.Context, id string, updates UpdateFunctionRequest) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	fn, ok := db.functions[id]
	if !ok {
		return ErrFunctionNotFound
	}

	if updates.Name != nil {
		fn.Name = *updates.Name
	}
	if updates.Description != nil {
		fn.Description = updates.Description
	}
	if updates.Disabled != nil {
		fn.Disabled = *updates.Disabled
	}
	if updates.RetentionDays != nil {
		fn.RetentionDays = updates.RetentionDays
	}
	if updates.CronSchedule != nil {
		fn.CronSchedule = updates.CronSchedule
	}
	if updates.CronStatus != nil {
		fn.CronStatus = updates.CronStatus
	}

	fn.UpdatedAt = time.Now().Unix()
	db.functions[id] = fn
	return nil
}

func (db *MemoryDB) DeleteFunction(_ context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.functions[id]; !ok {
		return ErrFunctionNotFound
	}

	delete(db.functions, id)
	delete(db.versions, id)
	return nil
}

// Version operations

func (db *MemoryDB) CreateVersion(_ context.Context, functionID string, code string, createdBy *string) (FunctionVersion, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.functions[functionID]; !ok {
		return FunctionVersion{}, ErrFunctionNotFound
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

func (db *MemoryDB) GetVersion(_ context.Context, functionID string, version int) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	versions := db.versions[functionID]
	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return FunctionVersion{}, ErrVersionNotFound
}

func (db *MemoryDB) GetVersionByID(_ context.Context, versionID string) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, versions := range db.versions {
		for _, v := range versions {
			if v.ID == versionID {
				return v, nil
			}
		}
	}

	return FunctionVersion{}, ErrVersionNotFound
}

func (db *MemoryDB) ListVersions(_ context.Context, functionID string, params PaginationParams) ([]FunctionVersion, int64, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Normalize pagination parameters
	params = params.Normalize()

	versions := db.versions[functionID]
	if versions == nil {
		return []FunctionVersion{}, 0, nil
	}

	// Return in reverse order (newest first)
	allVersions := make([]FunctionVersion, len(versions))
	for i, v := range versions {
		allVersions[len(versions)-1-i] = v
	}

	total := int64(len(allVersions))

	// Apply pagination
	start := params.Offset
	if start > len(allVersions) {
		return []FunctionVersion{}, total, nil
	}

	end := start + params.Limit
	if end > len(allVersions) {
		end = len(allVersions)
	}

	return allVersions[start:end], total, nil
}

func (db *MemoryDB) GetActiveVersion(_ context.Context, functionID string) (FunctionVersion, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	versions := db.versions[functionID]
	for _, v := range versions {
		if v.IsActive {
			return v, nil
		}
	}

	return FunctionVersion{}, ErrNoActiveVersion
}

func (db *MemoryDB) ActivateVersion(_ context.Context, functionID string, version int) error {
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
		return ErrVersionNotFound
	}

	db.versions[functionID] = versions
	return nil
}

// Execution operations

func (db *MemoryDB) CreateExecution(_ context.Context, exec Execution) (Execution, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Only set CreatedAt if not already set (allows manual timestamps for testing)
	if exec.CreatedAt == 0 {
		exec.CreatedAt = time.Now().Unix()
	}
	// Default trigger to HTTP if not set
	if exec.Trigger == "" {
		exec.Trigger = ExecutionTriggerHTTP
	}
	db.executions[exec.ID] = exec
	return exec, nil
}

func (db *MemoryDB) GetExecution(_ context.Context, executionID string) (Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, ok := db.executions[executionID]
	if !ok {
		return Execution{}, ErrExecutionNotFound
	}
	return exec, nil
}

func (db *MemoryDB) UpdateExecution(_ context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, ok := db.executions[executionID]
	if !ok {
		return ErrExecutionNotFound
	}

	exec.Status = status
	exec.DurationMs = durationMs
	exec.ErrorMessage = errorMsg
	db.executions[executionID] = exec

	return nil
}

func (db *MemoryDB) ListExecutions(_ context.Context, functionID string, params PaginationParams) ([]Execution, int64, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Normalize pagination parameters
	params = params.Normalize()

	var allExecutions []Execution
	for _, exec := range db.executions {
		if exec.FunctionID == functionID {
			allExecutions = append(allExecutions, exec)
		}
	}

	total := int64(len(allExecutions))

	// Apply pagination
	start := params.Offset
	if start > len(allExecutions) {
		return []Execution{}, total, nil
	}

	end := min(start+params.Limit, len(allExecutions))

	return allExecutions[start:end], total, nil
}

func (db *MemoryDB) DeleteOldExecutions(_ context.Context, beforeTimestamp int64) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var deletedCount int64
	for id, exec := range db.executions {
		if exec.CreatedAt < beforeTimestamp {
			delete(db.executions, id)
			deletedCount++
		}
	}

	return deletedCount, nil
}

func (db *MemoryDB) ListFunctionsWithActiveCron(_ context.Context) ([]Function, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var functions []Function
	for _, fn := range db.functions {
		if fn.CronStatus != nil && *fn.CronStatus == string(CronStatusActive) &&
			fn.CronSchedule != nil && *fn.CronSchedule != "" {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

// Health check

func (db *MemoryDB) Ping(_ context.Context) error {
	return nil
}
