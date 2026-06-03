// Package validation holds the input-validation rules for functions, versions,
// environment variables, and KV entries.
//
// It is transport-agnostic: both the REST handlers (internal/api) and the
// GraphQL resolvers (internal/graph) call into it, so the rules have a single
// home and cannot drift between the two APIs. When the REST layer is removed,
// validation continues to live here.
package validation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/robfig/cron/v3"
)

const (
	// MaxFunctionNameLength is the maximum length for function names.
	MaxFunctionNameLength = 100
	// MaxDescriptionLength is the maximum length for function descriptions.
	MaxDescriptionLength = 500
	// MaxCodeLength is the maximum length for function code.
	MaxCodeLength = 1024 * 1024 // 1MB
	// MaxEnvVarKeyLength is the maximum length for environment variable keys.
	MaxEnvVarKeyLength = 100
	// MaxEnvVarValueLength is the maximum length for environment variable values.
	MaxEnvVarValueLength = 10000
	// MaxEnvVars is the maximum number of environment variables per function.
	MaxEnvVars = 100
	// MaxStoreKeyLength is the maximum length for store keys.
	MaxStoreKeyLength = 100
	// MaxStoreValueLength is the maximum length for store values.
	MaxStoreValueLength = 10000
)

// AllowedRetentionDays lists the permitted execution-history retention windows.
var AllowedRetentionDays = []int{7, 15, 30, 365}

// AllowedCronStatuses lists the permitted cron schedule statuses.
var AllowedCronStatuses = []string{string(store.CronStatusActive), string(store.CronStatusPaused)}

// Error is a validation failure tied to a specific field.
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// CreateFunction validates the fields supplied when creating a function.
func CreateFunction(name string, description *string, code string) error {
	if err := FunctionName(name); err != nil {
		return err
	}
	if description != nil {
		if err := Description(*description); err != nil {
			return err
		}
	}
	return Code(code)
}

// UpdateFunctionRequest validates a function update. At least one field must be
// set, and any provided field must itself be valid.
func UpdateFunctionRequest(req *store.UpdateFunctionRequest) error {
	if req == nil {
		return &Error{Field: "request", Message: "request cannot be nil"}
	}

	if req.Name == nil && req.Description == nil && req.Code == nil && req.Disabled == nil && req.RetentionDays == nil && req.CronSchedule == nil && req.CronStatus == nil && req.SaveResponse == nil {
		return &Error{Field: "request", Message: "at least one field must be provided for update"}
	}

	if req.Name != nil {
		if err := FunctionName(*req.Name); err != nil {
			return err
		}
	}
	if req.Description != nil {
		if err := Description(*req.Description); err != nil {
			return err
		}
	}
	if req.Code != nil {
		if err := Code(*req.Code); err != nil {
			return err
		}
	}
	if req.RetentionDays != nil {
		if err := RetentionDays(*req.RetentionDays); err != nil {
			return err
		}
	}
	if req.CronSchedule != nil {
		if err := CronSchedule(*req.CronSchedule); err != nil {
			return err
		}
	}
	if req.CronStatus != nil {
		if err := CronStatus(*req.CronStatus); err != nil {
			return err
		}
	}
	return nil
}

// EnvVars validates a complete set of environment variables (count and each
// key/value). It does not reject an empty map.
func EnvVars(envVars map[string]string) error {
	if len(envVars) > MaxEnvVars {
		return &Error{
			Field:   "env_vars",
			Message: fmt.Sprintf("cannot have more than %d environment variables", MaxEnvVars),
		}
	}
	for key, value := range envVars {
		if err := EnvVarKey(key); err != nil {
			return err
		}
		if err := EnvVarValue(value); err != nil {
			return err
		}
	}
	return nil
}

// KVEntries validates a complete set of key/value store entries.
func KVEntries(entries map[string]string) error {
	for key, value := range entries {
		if err := StoreKey(key); err != nil {
			return err
		}
		if err := StoreValue(value); err != nil {
			return err
		}
	}
	return nil
}

// FunctionName validates a function name.
func FunctionName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return &Error{Field: "name", Message: "name cannot be empty"}
	}
	if len(trimmed) > MaxFunctionNameLength {
		return &Error{
			Field:   "name",
			Message: fmt.Sprintf("name cannot be longer than %d characters", MaxFunctionNameLength),
		}
	}
	return nil
}

// Description validates a function description.
func Description(description string) error {
	if len(description) > MaxDescriptionLength {
		return &Error{
			Field:   "description",
			Message: fmt.Sprintf("description cannot be longer than %d characters", MaxDescriptionLength),
		}
	}
	return nil
}

// Code validates function code.
func Code(code string) error {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return &Error{Field: "code", Message: "code cannot be empty"}
	}
	if len(code) > MaxCodeLength {
		return &Error{
			Field:   "code",
			Message: fmt.Sprintf("code cannot be longer than %d bytes", MaxCodeLength),
		}
	}
	return nil
}

// EnvVarKey validates an environment variable key.
func EnvVarKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return &Error{Field: "env_var_key", Message: "environment variable key cannot be empty"}
	}
	if len(key) > MaxEnvVarKeyLength {
		return &Error{
			Field:   "env_var_key",
			Message: fmt.Sprintf("environment variable key cannot be longer than %d characters", MaxEnvVarKeyLength),
		}
	}
	if !IsValidEnvVarKey(key) {
		return &Error{
			Field:   "env_var_key",
			Message: "environment variable key can only contain letters, numbers, and underscores",
		}
	}
	return nil
}

// EnvVarValue validates an environment variable value.
func EnvVarValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return &Error{Field: "env_var_value", Message: "environment variable value cannot be empty"}
	}
	if len(value) > MaxEnvVarValueLength {
		return &Error{
			Field:   "env_var_value",
			Message: fmt.Sprintf("environment variable value cannot be longer than %d characters", MaxEnvVarValueLength),
		}
	}
	return nil
}

// IsValidEnvVarKey reports whether a string is a valid environment variable key
// (letters, numbers, and underscores only).
func IsValidEnvVarKey(key string) bool {
	if key == "" {
		return false
	}
	for _, char := range key {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}

// RetentionDays validates an execution-history retention window.
func RetentionDays(days int) error {
	if slices.Contains(AllowedRetentionDays, days) {
		return nil
	}
	return &Error{
		Field:   "retention_days",
		Message: fmt.Sprintf("retention_days must be one of: %v", AllowedRetentionDays),
	}
}

// CronSchedule validates a cron expression. An empty schedule is allowed (it
// clears the schedule).
func CronSchedule(schedule string) error {
	if schedule == "" {
		return nil
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(schedule); err != nil {
		return &Error{
			Field:   "cron_schedule",
			Message: fmt.Sprintf("invalid cron expression: %v", err),
		}
	}
	return nil
}

// CronStatus validates a cron status value.
func CronStatus(status string) error {
	if slices.Contains(AllowedCronStatuses, status) {
		return nil
	}
	return &Error{
		Field:   "cron_status",
		Message: fmt.Sprintf("cron_status must be one of: %v", AllowedCronStatuses),
	}
}

// StoreKey validates a KV store key.
func StoreKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return &Error{Field: "key", Message: "key cannot be empty"}
	}
	if len(key) > MaxStoreKeyLength {
		return &Error{
			Field:   "key",
			Message: fmt.Sprintf("key cannot be longer than %d characters", MaxStoreKeyLength),
		}
	}
	return nil
}

// StoreValue validates a KV store value.
func StoreValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return &Error{Field: "value", Message: "value cannot be empty"}
	}
	if len(value) > MaxStoreValueLength {
		return &Error{
			Field:   "value",
			Message: fmt.Sprintf("value cannot be longer than %d characters", MaxStoreValueLength),
		}
	}
	return nil
}
