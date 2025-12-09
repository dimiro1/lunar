package cron

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimiro1/lunar/internal/store"
)

func strPtr(s string) *string {
	return &s
}

func TestGetNextRunFromSchedule(t *testing.T) {
	tests := []struct {
		name        string
		schedule    string
		expectError bool
		expectNil   bool
	}{
		{
			name:        "valid every minute",
			schedule:    "* * * * *",
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "valid every 5 minutes",
			schedule:    "*/5 * * * *",
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "valid every hour",
			schedule:    "0 * * * *",
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "valid specific time",
			schedule:    "30 14 * * *",
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "empty schedule",
			schedule:    "",
			expectError: false,
			expectNil:   true,
		},
		{
			name:        "invalid schedule - too few fields",
			schedule:    "* * *",
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "invalid schedule - bad syntax",
			schedule:    "not a cron",
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetNextRunFromSchedule(tt.schedule)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectNil && result != nil {
				t.Errorf("expected nil result but got %v", result)
			}
			if !tt.expectNil && result == nil {
				t.Errorf("expected non-nil result but got nil")
			}
			if result != nil && result.Before(time.Now()) {
				t.Errorf("next run should be in the future, got %v", result)
			}
		})
	}
}

func TestFormatNextRun(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "overdue",
			time:     now.Add(-1 * time.Hour),
			expected: "overdue",
		},
		{
			name:     "less than a minute",
			time:     now.Add(30 * time.Second),
			expected: "in less than a minute",
		},
		{
			name:     "1 minute",
			time:     now.Add(1*time.Minute + 30*time.Second),
			expected: "in 1 minute",
		},
		{
			name:     "5 minutes",
			time:     now.Add(5*time.Minute + 30*time.Second),
			expected: "in 5 minutes",
		},
		{
			name:     "1 hour",
			time:     now.Add(1*time.Hour + 30*time.Minute),
			expected: "in 1 hour",
		},
		{
			name:     "5 hours",
			time:     now.Add(5*time.Hour + 30*time.Minute),
			expected: "in 5 hours",
		},
		{
			name:     "1 day",
			time:     now.Add(25 * time.Hour),
			expected: "in 1 day",
		},
		{
			name:     "3 days",
			time:     now.Add(73 * time.Hour),
			expected: "in 3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNextRun(tt.time)
			if result != tt.expected {
				t.Errorf("FormatNextRun() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFunctionScheduler_NewScheduler(t *testing.T) {
	db := store.NewMemoryDB()
	scheduler := NewScheduler(db, "http://localhost:8080")

	if scheduler == nil {
		t.Fatal("NewScheduler returned nil")
	}
	if scheduler.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q, expected %q", scheduler.baseURL, "http://localhost:8080")
	}
	if scheduler.jobs == nil {
		t.Error("jobs map not initialized")
	}
	if scheduler.client == nil {
		t.Error("http client not initialized")
	}
}

func TestFunctionScheduler_StartStop(t *testing.T) {
	db := store.NewMemoryDB()
	scheduler := NewScheduler(db, "http://localhost:8080")

	// Start should succeed with empty DB
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Stop should complete without hanging
	done := make(chan struct{})
	go func() {
		scheduler.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Stop() timed out")
	}
}

func TestFunctionScheduler_LoadSchedules(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Add a function with active cron schedule
	_, err := db.CreateFunction(ctx, store.Function{
		ID:           "func-1",
		Name:         "test-function",
		CronSchedule: strPtr("*/5 * * * *"),
		CronStatus:   strPtr("active"),
	})
	if err != nil {
		t.Fatalf("CreateFunction() failed: %v", err)
	}

	scheduler := NewScheduler(db, "http://localhost:8080")
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer scheduler.Stop()

	// Check that the job was added
	scheduler.mu.RLock()
	_, exists := scheduler.jobs["func-1"]
	scheduler.mu.RUnlock()

	if !exists {
		t.Error("expected job to be added for func-1")
	}
}

func TestFunctionScheduler_RefreshFunction(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Add a function without cron initially
	_, err := db.CreateFunction(ctx, store.Function{
		ID:   "func-1",
		Name: "test-function",
	})
	if err != nil {
		t.Fatalf("CreateFunction() failed: %v", err)
	}

	scheduler := NewScheduler(db, "http://localhost:8080")
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer scheduler.Stop()

	// Initially no job
	scheduler.mu.RLock()
	_, exists := scheduler.jobs["func-1"]
	scheduler.mu.RUnlock()
	if exists {
		t.Error("expected no job initially")
	}

	// Update the function with cron schedule
	err = db.UpdateFunction(ctx, "func-1", store.UpdateFunctionRequest{
		CronSchedule: strPtr("*/5 * * * *"),
		CronStatus:   strPtr("active"),
	})
	if err != nil {
		t.Fatalf("UpdateFunction() failed: %v", err)
	}

	// Refresh
	if err := scheduler.RefreshFunction("func-1"); err != nil {
		t.Fatalf("RefreshFunction() failed: %v", err)
	}

	// Now job should exist
	scheduler.mu.RLock()
	_, exists = scheduler.jobs["func-1"]
	scheduler.mu.RUnlock()
	if !exists {
		t.Error("expected job to be added after refresh")
	}

	// Pause the schedule
	err = db.UpdateFunction(ctx, "func-1", store.UpdateFunctionRequest{
		CronStatus: strPtr("paused"),
	})
	if err != nil {
		t.Fatalf("UpdateFunction() failed: %v", err)
	}

	// Refresh again
	if err := scheduler.RefreshFunction("func-1"); err != nil {
		t.Fatalf("RefreshFunction() failed: %v", err)
	}

	// Job should be removed
	scheduler.mu.RLock()
	_, exists = scheduler.jobs["func-1"]
	scheduler.mu.RUnlock()
	if exists {
		t.Error("expected job to be removed after pausing")
	}
}

func TestFunctionScheduler_GetNextRun(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Add a function with active cron schedule
	_, err := db.CreateFunction(ctx, store.Function{
		ID:           "func-1",
		Name:         "test-function",
		CronSchedule: strPtr("* * * * *"), // Every minute
		CronStatus:   strPtr("active"),
	})
	if err != nil {
		t.Fatalf("CreateFunction() failed: %v", err)
	}

	scheduler := NewScheduler(db, "http://localhost:8080")
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer scheduler.Stop()

	// Get next run for scheduled function
	nextRun := scheduler.GetNextRun("func-1")
	if nextRun == nil {
		t.Fatal("expected next run time for func-1")
	}

	// Should be in the future
	if nextRun.Before(time.Now()) {
		t.Errorf("next run should be in the future, got %v", nextRun)
	}

	// Should be within the next minute (since schedule is every minute)
	if nextRun.After(time.Now().Add(2 * time.Minute)) {
		t.Errorf("next run should be within 2 minutes, got %v", nextRun)
	}

	// Get next run for non-existent function
	nextRunNil := scheduler.GetNextRun("non-existent")
	if nextRunNil != nil {
		t.Errorf("expected nil for non-existent function, got %v", nextRunNil)
	}
}

func TestFunctionScheduler_ExecuteFunction_Headers(t *testing.T) {
	var receivedHeaders http.Header
	var receivedMethod string
	var requestReceived bool

	// Create a test server to capture the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		receivedMethod = r.Method
		requestReceived = true
		w.Header().Set("X-Execution-Id", "test-exec-id")
		w.Header().Set("X-Execution-Duration-Ms", "100")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db := store.NewMemoryDB()
	ctx := context.Background()

	_, err := db.CreateFunction(ctx, store.Function{
		ID:           "func-1",
		Name:         "test-function",
		CronSchedule: strPtr("* * * * *"), // Every minute - but we'll call directly
		CronStatus:   strPtr("active"),
	})
	if err != nil {
		t.Fatalf("CreateFunction() failed: %v", err)
	}

	scheduler := NewScheduler(db, server.URL)

	// Directly call executeFunction to test headers
	scheduler.executeFunction("func-1", "test-function", "*/5 * * * *")

	if !requestReceived {
		t.Fatal("request was not received by test server")
	}

	// Check method
	if receivedMethod != http.MethodPost {
		t.Errorf("method = %q, expected POST", receivedMethod)
	}

	// Check headers
	if got := receivedHeaders.Get(HeaderTrigger); got != TriggerValueCron {
		t.Errorf("X-Trigger = %q, expected %q", got, TriggerValueCron)
	}
	if got := receivedHeaders.Get(HeaderCronSchedule); got != "*/5 * * * *" {
		t.Errorf("X-Cron-Schedule = %q, expected %q", got, "*/5 * * * *")
	}
	if got := receivedHeaders.Get(HeaderCronFunctionID); got != "func-1" {
		t.Errorf("X-Cron-Function-Id = %q, expected %q", got, "func-1")
	}
	if got := receivedHeaders.Get(HeaderCronFunctionName); got != "test-function" {
		t.Errorf("X-Cron-Function-Name = %q, expected %q", got, "test-function")
	}
	if got := receivedHeaders.Get(HeaderCronScheduledTime); got == "" {
		t.Error("X-Cron-Scheduled-Time header is missing")
	}
	if got := receivedHeaders.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, expected %q", got, "application/json")
	}
}

func TestFunctionScheduler_ExecuteFunction_ErrorHandling(t *testing.T) {
	// Test with a server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	db := store.NewMemoryDB()
	scheduler := NewScheduler(db, server.URL)

	// This should not panic and should log a warning
	scheduler.executeFunction("func-1", "test-function", "*/5 * * * *")

	// Test with invalid URL (connection refused)
	scheduler2 := NewScheduler(db, "http://localhost:1") // Invalid port
	scheduler2.executeFunction("func-1", "test-function", "*/5 * * * *")

	// Both should complete without panicking
}

func TestHeaderConstants(t *testing.T) {
	// Verify header constants are set correctly
	if HeaderTrigger != "X-Trigger" {
		t.Errorf("HeaderTrigger = %q, expected X-Trigger", HeaderTrigger)
	}
	if HeaderCronSchedule != "X-Cron-Schedule" {
		t.Errorf("HeaderCronSchedule = %q, expected X-Cron-Schedule", HeaderCronSchedule)
	}
	if HeaderCronFunctionID != "X-Cron-Function-Id" {
		t.Errorf("HeaderCronFunctionID = %q, expected X-Cron-Function-Id", HeaderCronFunctionID)
	}
	if HeaderCronFunctionName != "X-Cron-Function-Name" {
		t.Errorf("HeaderCronFunctionName = %q, expected X-Cron-Function-Name", HeaderCronFunctionName)
	}
	if HeaderCronScheduledTime != "X-Cron-Scheduled-Time" {
		t.Errorf("HeaderCronScheduledTime = %q, expected X-Cron-Scheduled-Time", HeaderCronScheduledTime)
	}
	if TriggerValueCron != "cron" {
		t.Errorf("TriggerValueCron = %q, expected cron", TriggerValueCron)
	}
}
