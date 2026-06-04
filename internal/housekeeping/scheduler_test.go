package housekeeping

import (
	"context"
	"testing"
	"time"

	"github.com/dimiro1/lunar/internal/store"
)

func TestNewScheduler(t *testing.T) {
	db := store.NewMemoryDB()
	scheduler := NewScheduler(db, 365)

	if scheduler == nil {
		t.Fatal("Expected scheduler to be created")
	}
	if scheduler.db == nil {
		t.Error("Expected scheduler.db to be set")
	}
	if scheduler.cron == nil {
		t.Error("Expected scheduler.cron to be set")
	}
}

func TestScheduler_CleanupOldExecutions(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Create function and version
	fn := store.Function{
		ID:      "func_cleanup_test",
		Name:    "cleanup-test",
		EnvVars: make(map[string]string),
	}

	retentionDays := 7
	fn.RetentionDays = &retentionDays

	created, err := db.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := db.CreateVersion(ctx, created.ID, "code", "", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create executions with different timestamps
	now := time.Now().Unix()
	oldTime := now - (10 * 24 * 60 * 60)   // 10 days ago (should be deleted)
	recentTime := now - (2 * 24 * 60 * 60) // 2 days ago (should remain)

	oldExec := store.Execution{
		ID:                "exec_old",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            store.ExecutionStatusSuccess,
		CreatedAt:         oldTime,
	}

	recentExec := store.Execution{
		ID:                "exec_recent",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            store.ExecutionStatusSuccess,
		CreatedAt:         recentTime,
	}

	_, err = db.CreateExecution(ctx, oldExec)
	if err != nil {
		t.Fatalf("CreateExecution (old) failed: %v", err)
	}

	_, err = db.CreateExecution(ctx, recentExec)
	if err != nil {
		t.Fatalf("CreateExecution (recent) failed: %v", err)
	}

	// Run cleanup
	scheduler := NewScheduler(db, 365)
	err = scheduler.cleanupOldExecutions(ctx)
	if err != nil {
		t.Fatalf("cleanupOldExecutions failed: %v", err)
	}

	// Verify old execution is deleted
	_, err = db.GetExecution(ctx, "exec_old")
	if err == nil {
		t.Error("Expected old execution to be deleted")
	}

	// Verify recent execution still exists
	_, err = db.GetExecution(ctx, "exec_recent")
	if err != nil {
		t.Errorf("Expected recent execution to still exist: %v", err)
	}
}

func TestScheduler_CleanupOldExecutions_DefaultRetention(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Create function WITHOUT explicit retention days (should use default 7)
	fn := store.Function{
		ID:      "func_default_retention",
		Name:    "default-retention-test",
		EnvVars: make(map[string]string),
	}

	created, err := db.CreateFunction(ctx, fn)
	if err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	ver, err := db.CreateVersion(ctx, created.ID, "code", "", nil)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Create execution 10 days old (should be deleted with default 7-day retention)
	now := time.Now().Unix()
	oldTime := now - (10 * 24 * 60 * 60)

	oldExec := store.Execution{
		ID:                "exec_old_default",
		FunctionID:        fn.ID,
		FunctionVersionID: ver.ID,
		Status:            store.ExecutionStatusSuccess,
		CreatedAt:         oldTime,
	}

	_, err = db.CreateExecution(ctx, oldExec)
	if err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
	}

	// Run cleanup
	scheduler := NewScheduler(db, 365)
	err = scheduler.cleanupOldExecutions(ctx)
	if err != nil {
		t.Fatalf("cleanupOldExecutions failed: %v", err)
	}

	// Verify old execution is deleted
	_, err = db.GetExecution(ctx, "exec_old_default")
	if err == nil {
		t.Error("Expected old execution to be deleted with default retention")
	}
}

func TestScheduler_CleanupOldExecutions_MultipleFunctions(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Create two functions with different retention periods
	fn1 := store.Function{
		ID:      "func_retention_7",
		Name:    "retention-7-test",
		EnvVars: make(map[string]string),
	}
	retention7 := 7
	fn1.RetentionDays = &retention7

	fn2 := store.Function{
		ID:      "func_retention_30",
		Name:    "retention-30-test",
		EnvVars: make(map[string]string),
	}
	retention30 := 30
	fn2.RetentionDays = &retention30

	_, err := db.CreateFunction(ctx, fn1)
	if err != nil {
		t.Fatalf("CreateFunction 1 failed: %v", err)
	}

	_, err = db.CreateFunction(ctx, fn2)
	if err != nil {
		t.Fatalf("CreateFunction 2 failed: %v", err)
	}

	ver1, err := db.CreateVersion(ctx, fn1.ID, "code1", "", nil)
	if err != nil {
		t.Fatalf("CreateVersion 1 failed: %v", err)
	}

	ver2, err := db.CreateVersion(ctx, fn2.ID, "code2", "", nil)
	if err != nil {
		t.Fatalf("CreateVersion 2 failed: %v", err)
	}

	// Create executions
	now := time.Now().Unix()
	time10DaysAgo := now - (10 * 24 * 60 * 60)
	time40DaysAgo := now - (40 * 24 * 60 * 60)

	// fn1 execution 10 days old - should remain (global cutoff uses longest retention)
	exec1 := store.Execution{
		ID:                "exec_fn1_recent",
		FunctionID:        fn1.ID,
		FunctionVersionID: ver1.ID,
		Status:            store.ExecutionStatusSuccess,
		CreatedAt:         time10DaysAgo,
	}

	// fn2 execution 40 days old - should be deleted (older than 30 days)
	exec2 := store.Execution{
		ID:                "exec_fn2_old",
		FunctionID:        fn2.ID,
		FunctionVersionID: ver2.ID,
		Status:            store.ExecutionStatusSuccess,
		CreatedAt:         time40DaysAgo,
	}

	_, err = db.CreateExecution(ctx, exec1)
	if err != nil {
		t.Fatalf("CreateExecution 1 failed: %v", err)
	}

	_, err = db.CreateExecution(ctx, exec2)
	if err != nil {
		t.Fatalf("CreateExecution 2 failed: %v", err)
	}

	// Run cleanup
	scheduler := NewScheduler(db, 365)
	err = scheduler.cleanupOldExecutions(ctx)
	if err != nil {
		t.Fatalf("cleanupOldExecutions failed: %v", err)
	}

	// Note: Current implementation uses the longest retention period (30 days) as global cutoff
	// This is a simplified approach since DeleteOldExecutions is global, not per-function

	// Verify fn1 execution still exists (10 days < 30 day global cutoff)
	_, err = db.GetExecution(ctx, "exec_fn1_recent")
	if err != nil {
		t.Errorf("Expected fn1 execution to still exist (within global cutoff): %v", err)
	}

	// Verify fn2 execution is deleted (40 days > 30 day retention)
	_, err = db.GetExecution(ctx, "exec_fn2_old")
	if err == nil {
		t.Error("Expected fn2 old execution to be deleted")
	}
}

func TestScheduler_CleanupOldExecutions_NoExecutions(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Run cleanup with no executions
	scheduler := NewScheduler(db, 365)
	err := scheduler.cleanupOldExecutions(ctx)
	if err != nil {
		t.Fatalf("cleanupOldExecutions should not fail with no executions: %v", err)
	}
}

func TestScheduler_StartAndStop(t *testing.T) {
	db := store.NewMemoryDB()
	scheduler := NewScheduler(db, 365)

	// Start scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop()

	// Test should complete without hanging
}

func TestDefaultRetentionDays(t *testing.T) {
	if DefaultRetentionDays != 7 {
		t.Errorf("Expected DefaultRetentionDays to be 7, got %d", DefaultRetentionDays)
	}
}
