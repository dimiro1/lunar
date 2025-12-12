package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dimiro1/lunar/internal/migrate"
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

	migrate.RunTest(t, db)

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
		Disabled:    false,
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
	if created.Disabled != false {
		t.Error("Expected Disabled to be false")
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
	v1, err := sqliteDB.CreateVersion(ctx, fn.ID, "v1", nil)
	if err != nil {
		t.Fatalf("CreateVersion v1 failed: %v", err)
	}
	if _, err := sqliteDB.CreateVersion(ctx, fn.ID, "v2", nil); err != nil {
		t.Fatalf("CreateVersion v2 failed: %v", err)
	}

	// Activate v1 using version ID
	if err := sqliteDB.ActivateVersion(ctx, v1.ID); err != nil {
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
	if err := sqliteDB.UpdateExecution(ctx, exec.ID, ExecutionStatusError, &duration, &errorMsg, nil); err != nil {
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

// Disabled function tests

func TestSQLiteDB_UpdateFunction_ToggleDisabled(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:       "func_disabled_test",
		Name:     "disabled-test",
		EnvVars:  make(map[string]string),
		Disabled: false,
	}

	created, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	if created.Disabled {
		t.Error("Expected Disabled to be false initially")
	}

	// Sleep to ensure UpdatedAt will be different
	time.Sleep(1 * time.Second)

	// Disable the function
	disabledTrue := true
	updates := UpdateFunctionRequest{
		Disabled: &disabledTrue,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Retrieve and verify
	updated, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if !updated.Disabled {
		t.Error("Expected Disabled to be true after update")
	}
	if updated.UpdatedAt <= created.UpdatedAt {
		t.Error("Expected UpdatedAt to be newer")
	}

	// Sleep again
	time.Sleep(1 * time.Second)

	// Enable the function again
	disabledFalse := false
	updates = UpdateFunctionRequest{
		Disabled: &disabledFalse,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Retrieve and verify
	enabled, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if enabled.Disabled {
		t.Error("Expected Disabled to be false after re-enabling")
	}
	if enabled.UpdatedAt <= updated.UpdatedAt {
		t.Error("Expected UpdatedAt to be newer after re-enabling")
	}
}

func TestSQLiteDB_GetFunction_DisabledField(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a disabled function
	fn := Function{
		ID:       "func_disabled_get",
		Name:     "disabled-get-test",
		EnvVars:  make(map[string]string),
		Disabled: true,
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

	if !retrieved.Disabled {
		t.Error("Expected Disabled to be true")
	}
}

func TestSQLiteDB_ListFunctions_IncludesDisabledStatus(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create one enabled and one disabled function
	fn1 := Function{
		ID:       "func_list_enabled",
		Name:     "enabled-func",
		EnvVars:  make(map[string]string),
		Disabled: false,
	}
	fn2 := Function{
		ID:       "func_list_disabled",
		Name:     "disabled-func",
		EnvVars:  make(map[string]string),
		Disabled: true,
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn1); err != nil {
		t.Fatalf("CreateFunction fn1 failed: %v", err)
	}
	if _, err := sqliteDB.CreateFunction(ctx, fn2); err != nil {
		t.Fatalf("CreateFunction fn2 failed: %v", err)
	}

	// List functions
	functions, total, err := sqliteDB.ListFunctions(ctx, PaginationParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListFunctions failed: %v", err)
	}

	if len(functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(functions))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	// Verify disabled status is preserved
	var foundEnabled, foundDisabled bool
	for _, fn := range functions {
		if fn.ID == fn1.ID && !fn.Disabled {
			foundEnabled = true
		}
		if fn.ID == fn2.ID && fn.Disabled {
			foundDisabled = true
		}
	}

	if !foundEnabled {
		t.Error("Expected to find enabled function with Disabled=false")
	}
	if !foundDisabled {
		t.Error("Expected to find disabled function with Disabled=true")
	}
}

// Event JSON tests

func TestSQLiteDB_CreateExecution_WithEventJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_event_json",
		Name:    "event-json-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create execution with event JSON
	eventJSON := `{"method":"POST","path":"/api/test","headers":{"Content-Type":"application/json"},"body":"test data","query":{}}`
	exec := Execution{
		ID:                "exec_with_event",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
		EventJSON:         &eventJSON,
	}

	created, err := sqliteDB.CreateExecution(ctx, exec)
	if err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	if created.ID != exec.ID {
		t.Errorf("Expected ID %s, got %s", exec.ID, created.ID)
	}
	if created.EventJSON == nil {
		t.Fatal("Expected EventJSON to be set")
	}
	if *created.EventJSON != eventJSON {
		t.Errorf("Expected EventJSON %s, got %s", eventJSON, *created.EventJSON)
	}
}

func TestSQLiteDB_GetExecution_WithEventJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and execution
	fn := Function{
		ID:      "func_get_event_json",
		Name:    "get-event-json-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	eventJSON := `{"method":"GET","path":"/api/users","headers":{"Authorization":"Bearer token"},"body":"","query":{"limit":"10"}}`
	exec := Execution{
		ID:                "exec_get_event",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
		EventJSON:         &eventJSON,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Get execution
	retrieved, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if retrieved.EventJSON == nil {
		t.Fatal("Expected EventJSON to be set")
	}
	if *retrieved.EventJSON != eventJSON {
		t.Errorf("Expected EventJSON %s, got %s", eventJSON, *retrieved.EventJSON)
	}
}

func TestSQLiteDB_GetExecution_WithoutEventJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and execution without event JSON
	fn := Function{
		ID:      "func_no_event_json",
		Name:    "no-event-json-test",
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
		ID:                "exec_no_event",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
		EventJSON:         nil, // No event JSON
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Get execution
	retrieved, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if retrieved.EventJSON != nil {
		t.Errorf("Expected EventJSON to be nil, got %v", *retrieved.EventJSON)
	}
}

func TestSQLiteDB_ListExecutions_WithEventJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_list_event_json",
		Name:    "list-event-json-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create multiple executions with event JSON
	eventJSON1 := `{"method":"POST","path":"/api/v1","headers":{},"body":"data1","query":{}}`
	eventJSON2 := `{"method":"GET","path":"/api/v2","headers":{},"body":"","query":{"id":"123"}}`

	exec1 := Execution{
		ID:                "exec_list_1",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
		EventJSON:         &eventJSON1,
	}

	exec2 := Execution{
		ID:                "exec_list_2",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
		EventJSON:         &eventJSON2,
	}

	// Create execution without event JSON
	exec3 := Execution{
		ID:                "exec_list_3",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
		EventJSON:         nil,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec1); err != nil {
		t.Fatalf("CreateExecution 1 failed: %v", err)
	}
	if _, err := sqliteDB.CreateExecution(ctx, exec2); err != nil {
		t.Fatalf("CreateExecution 2 failed: %v", err)
	}
	if _, err := sqliteDB.CreateExecution(ctx, exec3); err != nil {
		t.Fatalf("CreateExecution 3 failed: %v", err)
	}

	// List executions
	params := PaginationParams{Limit: 10, Offset: 0}
	executions, total, err := sqliteDB.ListExecutions(ctx, fn.ID, params)
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if len(executions) != 3 {
		t.Fatalf("Expected 3 executions, got %d", len(executions))
	}

	// Verify event JSON is included in results
	foundWithEvent := 0
	foundWithoutEvent := 0

	for _, exec := range executions {
		if exec.EventJSON != nil {
			foundWithEvent++
			// Verify the content matches
			if exec.ID == "exec_list_1" && *exec.EventJSON != eventJSON1 {
				t.Errorf("Expected EventJSON %s, got %s", eventJSON1, *exec.EventJSON)
			}
			if exec.ID == "exec_list_2" && *exec.EventJSON != eventJSON2 {
				t.Errorf("Expected EventJSON %s, got %s", eventJSON2, *exec.EventJSON)
			}
		} else {
			foundWithoutEvent++
		}
	}

	if foundWithEvent != 2 {
		t.Errorf("Expected 2 executions with EventJSON, got %d", foundWithEvent)
	}
	if foundWithoutEvent != 1 {
		t.Errorf("Expected 1 execution without EventJSON, got %d", foundWithoutEvent)
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

// Retention tests

func TestSQLiteDB_UpdateFunction_RetentionDays(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:      "func_retention",
		Name:    "retention-test",
		EnvVars: make(map[string]string),
	}

	created, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	if created.RetentionDays != nil {
		t.Error("Expected RetentionDays to be nil initially")
	}

	// Update retention days
	retentionDays := 30
	updates := UpdateFunctionRequest{
		RetentionDays: &retentionDays,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Retrieve and verify
	updated, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if updated.RetentionDays == nil {
		t.Fatal("Expected RetentionDays to be set")
	}
	if *updated.RetentionDays != 30 {
		t.Errorf("Expected RetentionDays 30, got %d", *updated.RetentionDays)
	}
}

func TestSQLiteDB_DeleteOldExecutions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_cleanup",
		Name:    "cleanup-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create executions with different timestamps
	now := time.Now().Unix()
	oldTime := now - (10 * 24 * 60 * 60)   // 10 days ago
	recentTime := now - (2 * 24 * 60 * 60) // 2 days ago

	// Manually insert executions with specific timestamps
	_, err = db.ExecContext(ctx, `INSERT INTO executions (id, function_id, function_version_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		"exec_old", fn.ID, ver.ID, ExecutionStatusSuccess, oldTime)
	if err != nil {
		t.Fatalf("Failed to insert old execution: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO executions (id, function_id, function_version_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		"exec_recent", fn.ID, ver.ID, ExecutionStatusSuccess, recentTime)
	if err != nil {
		t.Fatalf("Failed to insert recent execution: %v", err)
	}

	// Delete executions older than 7 days
	cutoffTime := now - (7 * 24 * 60 * 60)
	deleted, err := sqliteDB.DeleteOldExecutions(ctx, cutoffTime)
	if err != nil {
		t.Fatalf("DeleteOldExecutions failed: %v", err)
	}

	if deleted != 1 {
		t.Errorf("Expected 1 deleted execution, got %d", deleted)
	}

	// Verify old execution is gone
	_, err = sqliteDB.GetExecution(ctx, "exec_old")
	if err == nil {
		t.Error("Expected old execution to be deleted")
	}

	// Verify recent execution still exists
	_, err = sqliteDB.GetExecution(ctx, "exec_recent")
	if err != nil {
		t.Errorf("Expected recent execution to still exist: %v", err)
	}
}

func TestSQLiteDB_DeleteOldExecutions_NoExecutions(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Delete without any executions
	cutoffTime := time.Now().Unix()
	deleted, err := sqliteDB.DeleteOldExecutions(ctx, cutoffTime)
	if err != nil {
		t.Fatalf("DeleteOldExecutions failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("Expected 0 deleted executions, got %d", deleted)
	}
}

func TestSQLiteDB_DeleteOldExecutions_AllNew(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function, version, and recent execution
	fn := Function{
		ID:      "func_cleanup_new",
		Name:    "cleanup-new-test",
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
		ID:                "exec_new",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusSuccess,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Try to delete with cutoff time in the past (before execution was created)
	cutoffTime := time.Now().Unix() - (30 * 24 * 60 * 60) // 30 days ago
	deleted, err := sqliteDB.DeleteOldExecutions(ctx, cutoffTime)
	if err != nil {
		t.Fatalf("DeleteOldExecutions failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("Expected 0 deleted executions, got %d", deleted)
	}

	// Verify execution still exists
	_, err = sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Errorf("Expected execution to still exist: %v", err)
	}
}

// Save Response tests

func TestSQLiteDB_UpdateFunction_SaveResponse(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a function
	fn := Function{
		ID:           "func_save_response",
		Name:         "save-response-test",
		EnvVars:      make(map[string]string),
		SaveResponse: false,
	}

	created, err := sqliteDB.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	if created.SaveResponse {
		t.Error("Expected SaveResponse to be false initially")
	}

	// Enable save_response
	saveResponse := true
	updates := UpdateFunctionRequest{
		SaveResponse: &saveResponse,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Verify save_response is enabled
	updated, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if !updated.SaveResponse {
		t.Error("Expected SaveResponse to be true after update")
	}

	// Disable save_response
	saveResponseFalse := false
	updates2 := UpdateFunctionRequest{
		SaveResponse: &saveResponseFalse,
	}

	if err := sqliteDB.UpdateFunction(ctx, fn.ID, updates2); err != nil {
		t.Fatalf("UpdateFunction failed: %v", err)
	}

	// Verify save_response is disabled
	updated2, err := sqliteDB.GetFunction(ctx, fn.ID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}

	if updated2.SaveResponse {
		t.Error("Expected SaveResponse to be false after update")
	}
}

func TestSQLiteDB_UpdateExecution_WithResponseJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_response_json",
		Name:    "response-json-test",
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
		ID:                "exec_response_json",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Update execution with response JSON
	durationMs := int64(100)
	responseJSON := `{"statusCode":200,"headers":{"Content-Type":"application/json"},"body":"{\"success\":true}","isBase64Encoded":false}`

	if err := sqliteDB.UpdateExecution(ctx, exec.ID, ExecutionStatusSuccess, &durationMs, nil, &responseJSON); err != nil {
		t.Fatalf("UpdateExecution failed: %v", err)
	}

	// Verify response JSON was stored
	updated, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if updated.ResponseJSON == nil {
		t.Fatal("Expected ResponseJSON to be set")
	}

	if *updated.ResponseJSON != responseJSON {
		t.Errorf("Expected ResponseJSON %s, got %s", responseJSON, *updated.ResponseJSON)
	}

	if updated.Status != ExecutionStatusSuccess {
		t.Errorf("Expected status %s, got %s", ExecutionStatusSuccess, updated.Status)
	}

	if updated.DurationMs == nil || *updated.DurationMs != durationMs {
		t.Errorf("Expected duration %d, got %v", durationMs, updated.DurationMs)
	}
}

func TestSQLiteDB_GetExecution_WithResponseJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_get_response_json",
		Name:    "get-response-json-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create execution and update with response JSON
	exec := Execution{
		ID:                "exec_get_response",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	responseJSON := `{"statusCode":201,"headers":{"Location":"/items/123"},"body":"created","isBase64Encoded":false}`
	durationMs := int64(50)

	if err := sqliteDB.UpdateExecution(ctx, exec.ID, ExecutionStatusSuccess, &durationMs, nil, &responseJSON); err != nil {
		t.Fatalf("UpdateExecution failed: %v", err)
	}

	// Get execution and verify response JSON
	retrieved, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if retrieved.ResponseJSON == nil {
		t.Fatal("Expected ResponseJSON to be set")
	}

	if *retrieved.ResponseJSON != responseJSON {
		t.Errorf("Expected ResponseJSON %s, got %s", responseJSON, *retrieved.ResponseJSON)
	}
}

func TestSQLiteDB_GetExecution_WithoutResponseJSON(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create function and version
	fn := Function{
		ID:      "func_no_response_json",
		Name:    "no-response-json-test",
		EnvVars: make(map[string]string),
	}

	if _, err := sqliteDB.CreateFunction(ctx, fn); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := sqliteDB.CreateVersion(ctx, fn.ID, "code", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create execution and update without response JSON
	exec := Execution{
		ID:                "exec_no_response",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            ExecutionStatusPending,
	}

	if _, err := sqliteDB.CreateExecution(ctx, exec); err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	durationMs := int64(25)
	if err := sqliteDB.UpdateExecution(ctx, exec.ID, ExecutionStatusSuccess, &durationMs, nil, nil); err != nil {
		t.Fatalf("UpdateExecution failed: %v", err)
	}

	// Get execution and verify no response JSON
	retrieved, err := sqliteDB.GetExecution(ctx, exec.ID)
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if retrieved.ResponseJSON != nil {
		t.Errorf("Expected ResponseJSON to be nil, got %s", *retrieved.ResponseJSON)
	}
}
