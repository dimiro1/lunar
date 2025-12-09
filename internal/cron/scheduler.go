// Package cron provides scheduled execution of serverless functions using cron expressions.
//
// The scheduler executes functions via internal HTTP calls rather than direct invocation.
// This design was chosen to:
//   - Maintain compatibility with the existing function execution pipeline
//   - Reuse all middleware, logging, and metrics already in place
//   - Avoid duplicating execution logic and error handling
//   - Allow the scheduler to be stateless and easily replaceable
//   - Enable the same function to be triggered by both HTTP and cron without code changes
//
// When a cron job triggers, it makes an HTTP POST request to the function's endpoint
// with special headers to indicate the execution source. The function handler then
// records the execution with the appropriate trigger type.
package cron

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/robfig/cron/v3"
)

// Header constants for cron-triggered executions
const (
	// HeaderTrigger identifies the execution trigger source
	HeaderTrigger = "X-Trigger"
	// HeaderCronSchedule contains the cron expression that triggered the execution
	HeaderCronSchedule = "X-Cron-Schedule"
	// HeaderCronFunctionID contains the function ID being executed
	HeaderCronFunctionID = "X-Cron-Function-Id"
	// HeaderCronFunctionName contains the function name being executed
	HeaderCronFunctionName = "X-Cron-Function-Name"
	// HeaderCronScheduledTime contains the scheduled execution time (Unix timestamp)
	HeaderCronScheduledTime = "X-Cron-Scheduled-Time"

	// TriggerValueCron is the value for X-Trigger header when triggered by cron
	TriggerValueCron = "cron"
)

// FunctionScheduler manages cron schedules for serverless functions.
//
// It executes functions via internal HTTP calls when their schedules trigger.
// This approach maintains compatibility with the existing execution pipeline,
// reusing all middleware, logging, metrics, and error handling already in place.
//
// The scheduler is thread-safe and can be used concurrently. It maintains a mapping
// of function IDs to cron entry IDs to allow dynamic updates to schedules.
type FunctionScheduler struct {
	db      store.DB
	cron    *cron.Cron
	baseURL string
	jobs    map[string]cron.EntryID // functionID -> entryID
	mu      sync.RWMutex
	client  *http.Client
}

// NewScheduler creates a new function scheduler.
// baseURL should be the internal URL where functions can be invoked (e.g., "http://localhost:8080").
func NewScheduler(db store.DB, baseURL string) *FunctionScheduler {
	return &FunctionScheduler{
		db:      db,
		cron:    cron.New(),
		baseURL: baseURL,
		jobs:    make(map[string]cron.EntryID),
		client: &http.Client{
			Timeout: 5 * time.Minute, // Match execution timeout
		},
	}
}

// Start initializes and starts the scheduler.
// It loads all functions with active cron schedules and begins scheduling them.
func (s *FunctionScheduler) Start() error {
	if err := s.loadSchedules(); err != nil {
		return fmt.Errorf("failed to load schedules: %w", err)
	}

	s.cron.Start()
	slog.Info("Function cron scheduler started")
	return nil
}

// Stop gracefully stops the scheduler.
func (s *FunctionScheduler) Stop() {
	slog.Info("Stopping function cron scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("Function cron scheduler stopped")
}

// loadSchedules loads all active cron schedules from the database.
func (s *FunctionScheduler) loadSchedules() error {
	ctx := context.Background()
	functions, err := s.db.ListFunctionsWithActiveCron(ctx)
	if err != nil {
		return fmt.Errorf("failed to list functions with active cron: %w", err)
	}

	for _, fn := range functions {
		if err := s.addJob(fn); err != nil {
			slog.Error("Failed to add cron job for function",
				"function_id", fn.ID,
				"function_name", fn.Name,
				"error", err)
		}
	}

	slog.Info("Loaded cron schedules", "count", len(functions))
	return nil
}

// RefreshFunction updates the cron schedule for a specific function.
// Call this after updating a function's cron settings.
func (s *FunctionScheduler) RefreshFunction(functionID string) error {
	ctx := context.Background()
	fn, err := s.db.GetFunction(ctx, functionID)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job if any
	if entryID, exists := s.jobs[functionID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, functionID)
		slog.Info("Removed cron job for function", "function_id", functionID)
	}

	// Add new job if cron is active and has a schedule
	if fn.CronStatus != nil && *fn.CronStatus == string(store.CronStatusActive) &&
		fn.CronSchedule != nil && *fn.CronSchedule != "" {
		if err := s.addJobLocked(fn); err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
	}

	return nil
}

// addJob adds a cron job for a function (acquires lock).
func (s *FunctionScheduler) addJob(fn store.Function) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addJobLocked(fn)
}

// addJobLocked adds a cron job for a function (caller must hold lock).
func (s *FunctionScheduler) addJobLocked(fn store.Function) error {
	if fn.CronSchedule == nil || *fn.CronSchedule == "" {
		return nil
	}

	schedule := *fn.CronSchedule
	functionID := fn.ID
	functionName := fn.Name

	entryID, err := s.cron.AddFunc(schedule, func() {
		s.executeFunction(functionID, functionName, schedule)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.jobs[functionID] = entryID
	slog.Info("Added cron job for function",
		"function_id", functionID,
		"function_name", functionName,
		"schedule", schedule)

	return nil
}

// executeFunction triggers a function execution via internal HTTP call.
//
// This method uses HTTP to execute functions rather than direct invocation for several reasons:
//   - Maintains compatibility with the existing function execution pipeline
//   - Reuses all middleware, logging, metrics, and error handling already in place
//   - Avoids duplicating execution logic and the complexity of managing runner dependencies
//   - Allows the scheduler to remain stateless and easily replaceable
//   - Enables the same function code to work identically whether triggered by HTTP or cron
//
// The following headers are included in the request to provide context to the execution:
//   - X-Trigger: "cron" - indicates this is a cron-triggered execution
//   - X-Cron-Schedule: the cron expression that triggered the execution
//   - X-Cron-Function-Id: the function ID being executed
//   - X-Cron-Function-Name: the function name being executed
//   - X-Cron-Scheduled-Time: Unix timestamp of when the execution was scheduled
func (s *FunctionScheduler) executeFunction(functionID, functionName, schedule string) {
	scheduledTime := time.Now()
	url := fmt.Sprintf("%s/fn/%s", s.baseURL, functionID)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		slog.Error("Failed to create request for cron execution",
			"function_id", functionID,
			"function_name", functionName,
			"schedule", schedule,
			"error", err)
		return
	}

	// Add headers to provide context about the cron-triggered execution
	req.Header.Set(HeaderTrigger, TriggerValueCron)
	req.Header.Set(HeaderCronSchedule, schedule)
	req.Header.Set(HeaderCronFunctionID, functionID)
	req.Header.Set(HeaderCronFunctionName, functionName)
	req.Header.Set(HeaderCronScheduledTime, fmt.Sprintf("%d", scheduledTime.Unix()))
	req.Header.Set("Content-Type", "application/json")

	slog.Info("Executing function via cron",
		"function_id", functionID,
		"function_name", functionName,
		"schedule", schedule,
		"scheduled_time", scheduledTime.Format(time.RFC3339))

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Error("Cron execution failed",
			"function_id", functionID,
			"function_name", functionName,
			"schedule", schedule,
			"error", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	executionID := resp.Header.Get("X-Execution-Id")
	durationMs := resp.Header.Get("X-Execution-Duration-Ms")

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("Cron execution completed successfully",
			"function_id", functionID,
			"function_name", functionName,
			"schedule", schedule,
			"execution_id", executionID,
			"duration_ms", durationMs,
			"status_code", resp.StatusCode)
	} else {
		slog.Warn("Cron execution returned non-success status",
			"function_id", functionID,
			"function_name", functionName,
			"schedule", schedule,
			"execution_id", executionID,
			"duration_ms", durationMs,
			"status_code", resp.StatusCode)
	}
}

// GetNextRun calculates the next scheduled run time for a function.
// Returns nil if the function has no active schedule.
func (s *FunctionScheduler) GetNextRun(functionID string) *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, exists := s.jobs[functionID]
	if !exists {
		return nil
	}

	entry := s.cron.Entry(entryID)
	if entry.ID == 0 {
		return nil
	}

	next := entry.Next
	return &next
}

// GetNextRunFromSchedule calculates the next run time from a cron expression.
// This is useful for calculating next run without requiring an active job.
func GetNextRunFromSchedule(schedule string) (*time.Time, error) {
	if schedule == "" {
		return nil, nil
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	next := sched.Next(time.Now())
	return &next, nil
}

// FormatNextRun formats a time as a human-friendly relative string.
func FormatNextRun(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		return "overdue"
	}

	if diff < time.Minute {
		return "in less than a minute"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	}

	days := int(diff.Hours() / 24)
	if days == 1 {
		return "in 1 day"
	}
	return fmt.Sprintf("in %d days", days)
}
