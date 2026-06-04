package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

var _ DB = (*MemoryDB)(nil)

// MemoryDB is an in-memory implementation of the DB interface
type MemoryDB struct {
	mu         sync.RWMutex
	functions  map[string]Function
	versions   map[string][]FunctionVersion      // functionID -> versions
	executions map[string]Execution              // id -> execution
	apiTokens  map[string]APIToken               // id -> token
	metrics    map[string]map[int64]MetricBucket // functionID -> bucketHour -> bucket
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		functions:  make(map[string]Function),
		versions:   make(map[string][]FunctionVersion),
		executions: make(map[string]Execution),
		apiTokens:  make(map[string]APIToken),
		metrics:    make(map[string]map[int64]MetricBucket),
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

		// Find an active version
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

	end := min(start+params.Limit, len(allFunctions))

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
	if updates.SaveResponse != nil {
		fn.SaveResponse = *updates.SaveResponse
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

func (db *MemoryDB) CreateVersion(_ context.Context, functionID string, code string, language string, createdBy *string) (FunctionVersion, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.functions[functionID]; !ok {
		return FunctionVersion{}, ErrFunctionNotFound
	}

	versions := db.versions[functionID]
	versionNum := len(versions) + 1

	// Language is chosen at creation and sticky thereafter: when not specified,
	// inherit it from the function's most recent version (Lua for the first one).
	lang := Language(language)
	if lang == "" {
		if n := len(versions); n > 0 && versions[n-1].Language != "" {
			lang = versions[n-1].Language
		} else {
			lang = defaultLanguage
		}
	}

	// Deactivate all previous versions
	for i := range versions {
		versions[i].IsActive = false
	}

	version := FunctionVersion{
		ID:         fmt.Sprintf("ver_%s_v%d", functionID, versionNum),
		FunctionID: functionID,
		Version:    versionNum,
		Code:       code,
		Language:   lang,
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

	end := min(start+params.Limit, len(allVersions))

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

func (db *MemoryDB) ActivateVersion(_ context.Context, versionID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Find the version by ID across all functions
	var targetFunctionID string
	var targetIdx = -1

	for funcID, versions := range db.versions {
		for i, v := range versions {
			if v.ID == versionID {
				targetFunctionID = funcID
				targetIdx = i
				break
			}
		}
		if targetIdx != -1 {
			break
		}
	}

	if targetIdx == -1 {
		return ErrVersionNotFound
	}

	// Deactivate all versions and activate the target
	versions := db.versions[targetFunctionID]
	for i := range versions {
		versions[i].IsActive = (i == targetIdx)
	}

	db.versions[targetFunctionID] = versions
	return nil
}

func (db *MemoryDB) DeleteVersion(_ context.Context, versionID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Find the version by ID across all functions
	var targetFunctionID string
	var targetIdx = -1

	for funcID, versions := range db.versions {
		for i, v := range versions {
			if v.ID == versionID {
				// Check if it's the active version
				if v.IsActive {
					return ErrCannotDeleteActiveVersion
				}
				targetFunctionID = funcID
				targetIdx = i
				break
			}
		}
		if targetIdx != -1 {
			break
		}
	}

	if targetIdx == -1 {
		return ErrVersionNotFound
	}

	// Remove the version from the slice
	versions := db.versions[targetFunctionID]
	db.versions[targetFunctionID] = append(versions[:targetIdx], versions[targetIdx+1:]...)
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

func (db *MemoryDB) UpdateExecution(_ context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string, responseJSON *string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, ok := db.executions[executionID]
	if !ok {
		return ErrExecutionNotFound
	}

	exec.Status = status
	exec.DurationMs = durationMs
	exec.ErrorMessage = errorMsg
	exec.ResponseJSON = responseJSON
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

// Metric operations

func (db *MemoryDB) IncrementMetricBucket(_ context.Context, functionID string, bucketHour int64, isError bool, durationMs int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	buckets := db.metrics[functionID]
	if buckets == nil {
		buckets = make(map[int64]MetricBucket)
		db.metrics[functionID] = buckets
	}

	b := buckets[bucketHour]
	b.BucketStart = bucketHour
	b.Count++
	if isError {
		b.ErrorCount++
	}
	b.SumDurationMs += durationMs
	if durationMs > b.MaxDurationMs {
		b.MaxDurationMs = durationMs
	}
	buckets[bucketHour] = b
	return nil
}

func (db *MemoryDB) GetFunctionMetrics(_ context.Context, functionID string, fromUnix, toUnix, bucketSeconds int64) ([]MetricBucket, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if bucketSeconds <= 0 {
		bucketSeconds = 3600
	}

	// Group the stored hourly buckets into windows of bucketSeconds.
	grouped := make(map[int64]MetricBucket)
	for hour, b := range db.metrics[functionID] {
		if hour < fromUnix || hour >= toUnix {
			continue
		}
		start := (hour / bucketSeconds) * bucketSeconds
		g := grouped[start]
		g.BucketStart = start
		g.Count += b.Count
		g.ErrorCount += b.ErrorCount
		g.SumDurationMs += b.SumDurationMs
		if b.MaxDurationMs > g.MaxDurationMs {
			g.MaxDurationMs = b.MaxDurationMs
		}
		grouped[start] = g
	}

	out := make([]MetricBucket, 0, len(grouped))
	for _, b := range grouped {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].BucketStart < out[j].BucketStart })
	return out, nil
}

func (db *MemoryDB) DeleteOldMetricBuckets(_ context.Context, beforeBucketHour int64) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var deleted int64
	for functionID, buckets := range db.metrics {
		for hour := range buckets {
			if hour < beforeBucketHour {
				delete(buckets, hour)
				deleted++
			}
		}
		if len(buckets) == 0 {
			delete(db.metrics, functionID)
		}
	}
	return deleted, nil
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

// API Token operations

func (db *MemoryDB) CreateAPIToken(_ context.Context, token APIToken) (APIToken, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	token.CreatedAt = time.Now().Unix()
	db.apiTokens[token.ID] = token
	return token, nil
}

func (db *MemoryDB) GetAPITokenByHash(_ context.Context, tokenHash string) (APIToken, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, token := range db.apiTokens {
		if token.TokenHash == tokenHash && !token.Revoked {
			return token, nil
		}
	}

	return APIToken{}, ErrAPITokenNotFound
}

func (db *MemoryDB) ListAPITokens(_ context.Context) ([]APIToken, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	tokens := make([]APIToken, 0, len(db.apiTokens))
	for _, token := range db.apiTokens {
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (db *MemoryDB) RevokeAPIToken(_ context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	token, ok := db.apiTokens[id]
	if !ok {
		return ErrAPITokenNotFound
	}

	token.Revoked = true
	db.apiTokens[id] = token
	return nil
}

func (db *MemoryDB) UpdateAPITokenLastUsed(_ context.Context, id string, timestamp int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	token, ok := db.apiTokens[id]
	if !ok {
		return ErrAPITokenNotFound
	}

	token.LastUsed = &timestamp
	db.apiTokens[id] = token
	return nil
}

// Health check

func (db *MemoryDB) Ping(_ context.Context) error {
	return nil
}
