package api

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a new in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, *SQLiteDB) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db, NewSQLiteDB(db)
}

// Function operations tests

func TestSQLiteDB_CreateFunction(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	desc := "Test function"
	fn := Function{
		ID:          "func_test_123",
		Name:        "test-function",
		Description: &desc,
		EnvVars:     map[string]string{"KEY": "value"},
	}

	created, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	if created.ID != fn.ID {
		t.Errorf("Expected ID %s, got %s", fn.ID, created.ID)
	}
	if created.Name != fn.Name {
		t.Errorf("Expected Name %s, got %s", fn.Name, created.Name)
	}
	if created.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
	if created.UpdatedAt == 0 {
		t.Error("Expected UpdatedAt to be set")
	}
	if created.EnvVars["KEY"] != "value" {
		t.Error("Expected EnvVars to be preserved")
	}
}

func TestSQLiteDB_GetFunction(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_get_123",
		Name:    "get-test",
		EnvVars: map[string]string{"API_KEY": "secret"},
	}

	_, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Get the function
	retrieved, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if retrieved.ID != fn.ID {
		t.Errorf("Expected ID %s, got %s", fn.ID, retrieved.ID)
	}
	if retrieved.Name != fn.Name {
		t.Errorf("Expected Name %s, got %s", fn.Name, retrieved.Name)
	}
	if retrieved.EnvVars["API_KEY"] != "secret" {
		t.Error("Expected EnvVars to be retrieved correctly")
	}
}

func TestSQLiteDB_GetFunction_NotFound(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	_, err := sqliteDB.GetFunction(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent function")
	}
}

func TestSQLiteDB_ListFunctions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create multiple functions
	for i := range 3 {
		fn := Function{
			ID:      "func_list_" + string(rune('1'+i)),
			Name:    "list-test-" + string(rune('1'+i)),
			EnvVars: make(map[string]string),
		}
		if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
			t.Fatalf("CreateFunction failed: %v", err)
		}
	}

	functions, total, err := sqliteDB.ListFunctions(ctx, PaginationParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListFunctions failed: %v", err)
	}

	if len(functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(functions))
	}
	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}

func TestSQLiteDB_UpdateFunction(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_update_123",
		Name:    "original-name",
		EnvVars: make(map[string]string),
	}

	created, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Sleep to ensure UpdatedAt will be different (Unix timestamps are in seconds)
	time.Sleep(1 * time.Second)

	// Update the function
	newName := "updated-name"
	newDesc := "Updated description"
	updates := UpdateFunctionRequest{
		Name:        &newName,
		Description: &newDesc,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Retrieve and verify
	updated, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Expected Name %s, got %s", newName, updated.Name)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Error("Expected Description to be updated")
	}
	if updated.UpdatedAt <= created.UpdatedAt {
		t.Error("Expected UpdatedAt to be newer")
	}
}

func TestSQLiteDB_DeleteFunction(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_delete_123",
		Name:    "delete-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Delete the function
	if err := sqliteDB.DeleteFunction(ctx, fn.ID); err != nil {
		t.Fatalf("DeleteFunction failed: %v", err)
	}

	// Verify it's deleted
	_, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err == nil {
		t.Error("Expected error when getting deleted function")
	}
}

func TestSQLiteDB_UpdateFunctionEnvVars(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_env_123",
		Name:    "env-test",
		EnvVars: map[string]string{"OLD_KEY": "old_value"},
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Update env vars
	newEnvVars := map[string]string{
		"NEW_KEY": "new_value",
		"API_URL": "https://example.com",
	}

	if err := sqliteDB.UpdateFunctionEnvVars(ctx, fn.ID, newEnvVars); err != nil {
		t.Fatalf("UpdateFunctionEnvVars failed: %v", err)
	}

	// Verify
	updated, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if updated.EnvVars["NEW_KEY"] != "new_value" {
		t.Error("Expected NEW_KEY to be set")
	}
	if updated.EnvVars["API_URL"] != "https://example.com" {
		t.Error("Expected API_URL to be set")
	}
	if _, exists := updated.EnvVars["OLD_KEY"]; exists {
		t.Error("Expected OLD_KEY to be removed")
	}
}

// Version operations tests

func TestSQLiteDB_CreateVersion(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function first
	fn := Function{
		ID:      "func_ver_123",
		Name:    "version-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Create a version
	createdBy := "user@example.com"
	version, err := sqliteDB.CreateVersion(ctx, fn.ID, "function handler() end", &createdBy)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	if version.FunctionID != fn.ID {
		t.Errorf("Expected FunctionID %s, got %s", fn.ID, version.FunctionID)
	}
	if version.Version != 1 {
		t.Errorf("Expected Version 1, got %d", version.Version)
	}
	if !version.IsActive {
		t.Error("Expected first version to be active")
	}
	if version.CreatedBy == nil || *version.CreatedBy != createdBy {
		t.Error("Expected CreatedBy to be set")
	}
}

func TestSQLiteDB_CreateVersion_DeactivatesPrevious(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_ver_multi",
		Name:    "multi-version-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Create first version
	v1, err := sqliteDB.CreateVersion(ctx, fn.ID, "version 1", nil)
	if err != nil {
		t.Fatalf("CreateVersion v1 failed: %v", err)
	}

	// Create second version
	v2, err := sqliteDB.CreateVersion(ctx, fn.ID, "version 2", nil)
	if err != nil {
		t.Fatalf("CreateVersion v2 failed: %v", err)
	}

	// Verify v1 is inactive
	v1Retrieved, err := sqliteDB.GetVersion(ctx, fn.ID, v1.Version)
	if err != nil {
		t.Fatalf("GetVersion v1 failed: %v", err)
	}
	if v1Retrieved.IsActive {
		t.Error("Expected v1 to be inactive")
	}

	// Verify v2 is active
	if !v2.IsActive {
		t.Error("Expected v2 to be active")
	}
}

func TestSQLiteDB_GetVersion(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_getver",
		Name:    "getver-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	code := "function handler() return 42 end"
	created, err := sqliteDB.CreateVersion(ctx, fn.ID, code, nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Get the version
	retrieved, err := sqliteDB.GetVersion(ctx, fn.ID, 1)
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}
	if retrieved.Code != code {
		t.Errorf("Expected Code %s, got %s", code, retrieved.Code)
	}
}

func TestSQLiteDB_GetVersionByID(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_getverid",
		Name:    "getverid-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	created, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Get by ID
	retrieved, err := sqliteDB.GetVersionByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersionByID failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestSQLiteDB_ListVersions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function
	fn := Function{
		ID:      "func_listver",
		Name:    "listver-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Create 3 versions
	for i := 1; i <= 3; i++ {
		if _, err := sqliteDB.CreateVersion(ctx, fn.ID, "code v"+string(rune('0'+i)), nil); err != nil {
			t.Fatalf("CreateVersion v%d failed: %v", i, err)
		}
	}

	// List versions
	versions, total, err := sqliteDB.ListVersions(ctx, fn.ID, PaginationParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}
	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	// Verify they're in descending order
	if versions[0].Version != 3 || versions[1].Version != 2 || versions[2].Version != 1 {
		t.Error("Expected versions in descending order")
	}
}

func TestSQLiteDB_GetActiveVersion(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function
	fn := Function{
		ID:      "func_activever",
		Name:    "activever-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Create versions
	if _, err := sqliteDB.CreateVersion(ctx, fn.ID, "v1", nil); err != nil {
		t.Fatalf("CreateVersion v1 failed: %v", err)
	}
	v2, err := sqliteDB.CreateVersion(ctx, fn.ID, "v2", nil)
	if err != nil {
		t.Fatalf("CreateVersion v2 failed: %v", err)
	}

	// Get active version
	active, err := sqliteDB.GetActiveVersion(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetActiveVersion failed: %v", err)
	}

	if active.ID != v2.ID {
		t.Errorf("Expected active version ID %s, got %s", v2.ID, active.ID)
	}
	if active.Version != 2 {
		t.Errorf("Expected active version 2, got %d", active.Version)
	}
}

func TestSQLiteDB_ActivateVersion(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function
	fn := Function{
		ID:      "func_activate",
		Name:    "activate-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// Create 2 versions
	if _, err := sqliteDB.CreateVersion(ctx, fn.ID, "v1", nil); err != nil {
		t.Fatalf("CreateVersion v1 failed: %v", err)
	}
	if _, err := sqliteDB.CreateVersion(ctx, fn.ID, "v2", nil); err != nil {
		t.Fatalf("CreateVersion v2 failed: %v", err)
	}

	// Activate v1
	if err := sqliteDB.ActivateVersion(ctx, fn.ID, 1); err != nil {
		t.Fatalf("ActivateVersion failed: %v", err)
	}

	// Verify v1 is active
	active, err := sqliteDB.GetActiveVersion(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetActiveVersion failed: %v", err)
	}

	if active.Version != 1 {
		t.Errorf("Expected active version 1, got %d", active.Version)
	}
}

// Execution operations tests

func TestSQLiteDB_CreateExecution(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_exec",
		Name:    "exec-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create execution
	exec := Execution{
		ID:                "exec_123",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	created, err := sqliteDB.CreateExecution(ctx, exec)
	if err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	if created.ID != exec.ID {
		t.Errorf("Expected ID %s, got %s", exec.ID, created.ID)
	}
	if created.Status != ExecutionStatusPending {
		t.Errorf("Expected Status %s, got %s", ExecutionStatusPending, created.Status)
	}
	if created.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestSQLiteDB_GetExecution(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and execution
	fn := Function{
		ID:      "func_getexec",
		Name:    "getexec-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	exec := Execution{
		ID:                "exec_get",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Get execution
	retrieved, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if retrieved.ID != exec.ID {
		t.Errorf("Expected ID %s, got %s", exec.ID, retrieved.ID)
	}
	if retrieved.Status != ExecutionStatusPending {
		t.Errorf("Expected Status %s, got %s", ExecutionStatusPending, retrieved.Status)
	}
}

func TestSQLiteDB_UpdateExecution(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and execution
	fn := Function{
		ID:      "func_updateexec",
		Name:    "updateexec-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	exec := Execution{
		ID:                "exec_update",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Update execution
	duration := int64(250)
	errorMsg := "test error"
	if err := sqliteDB.UpdateExecution(ctx, exec.ID, ExecutionStatusError, &duration, &errorMsg); err != nil {
		t.Fatalf("UpdateExecution failed: %v", err)
	}

	// Verify
	updated, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if updated.Status != ExecutionStatusError {
		t.Errorf("Expected Status %s, got %s", ExecutionStatusError, updated.Status)
	}
	if updated.DurationMs == nil || *updated.DurationMs != duration {
		t.Error("Expected DurationMs to be set")
	}
	if updated.ErrorMessage == nil || *updated.ErrorMessage != errorMsg {
		t.Error("Expected ErrorMessage to be set")
	}
}

func TestSQLiteDB_ListExecutions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_listexec",
		Name:    "listexec-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create 3 executions
	for i := 1; i <= 3; i++ {
		exec := Execution{
			ID:                "exec_list_" + string(rune('0'+i)),
			FunctionID:        fn.ID,
			FunctionVersionID: ver.ID,
			Status:            ExecutionStatusPending,
		}
		if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
			t.Fatalf("CreateExecution %d failed: %v", i, err)
		}
	}

	// List executions
	executions, total, err := sqliteDB.ListExecutions(ctx, fn.ID, PaginationParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if len(executions) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executions))
	}
	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}

// CASCADE delete tests

func TestSQLiteDB_DeleteFunction_CascadesVersions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function with version
	fn := Function{
		ID:      "func_cascade",
		Name:    "cascade-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Delete function
	if err := sqliteDB.DeleteFunction(ctx, fn.ID); err != nil {
		t.Fatalf("DeleteFunction failed: %v", err)
	}

	// Verify version is also deleted
	_, err = sqliteDB.GetVersionByID(ctx, ver.ID)
	if err == nil {
		t.Error("Expected version to be cascade deleted")
	}
}

func TestSQLiteDB_DeleteFunction_CascadesExecutions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and execution
	fn := Function{
		ID:      "func_cascade_exec",
		Name:    "cascade-exec-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	exec := Execution{
		ID:                "exec_cascade",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Delete function
	if err := sqliteDB.DeleteFunction(ctx, fn.ID); err != nil {
		t.Fatalf("DeleteFunction failed: %v", err)
	}

	// Verify execution is also deleted
	_, err = sqliteDB.GetExecution(ctx, exec.ID)
	if err == nil {
		t.Error("Expected execution to be cascade deleted")
	}
}

// Health check test

func TestSQLiteDB_Ping(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	if err := sqliteDB.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}
