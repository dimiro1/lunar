package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dimiro1/lunar/internal/store"
	"github.com/robfig/cron/v3"
)

const (
	// MaxPageSize is the maximum allowed page size for pagination
	MaxPageSize = 100
	// MaxFunctionNameLength is the maximum length for function names
	MaxFunctionNameLength = 100
	// MaxDescriptionLength is the maximum length for function descriptions
	MaxDescriptionLength = 500
	// MaxCodeLength is the maximum length for function code
	MaxCodeLength = 1024 * 1024 // 1MB
	// MaxEnvVarKeyLength is the maximum length for environment variable keys
	MaxEnvVarKeyLength = 100
	// MaxEnvVarValueLength is the maximum length for environment variable values
	MaxEnvVarValueLength = 10000
	// MaxEnvVars is the maximum number of environment variables per function
	MaxEnvVars = 100
)

var AllowedRetentionDays = []int{7, 15, 30, 365}
var AllowedCronStatuses = []string{string(store.CronStatusActive), string(store.CronStatusPaused)}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateCreateFunctionRequest validates a CreateFunctionRequest
func ValidateCreateFunctionRequest(req *CreateFunctionRequest) error {
	if req == nil {
		return &ValidationError{Field: "request", Message: "request cannot be nil"}
	}

	// Validate name
	if err := validateFunctionName(req.Name); err != nil {
		return err
	}

	// Validate description if provided
	if req.Description != nil {
		if err := validateDescription(*req.Description); err != nil {
			return err
		}
	}

	// Validate code
	if err := validateCode(req.Code); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateFunctionRequest validates an UpdateFunctionRequest
func ValidateUpdateFunctionRequest(req *store.UpdateFunctionRequest) error {
	if req == nil {
		return &ValidationError{Field: "request", Message: "request cannot be nil"}
	}

	// At least one field must be provided
	if req.Name == nil && req.Description == nil && req.Code == nil && req.Disabled == nil && req.RetentionDays == nil && req.CronSchedule == nil && req.CronStatus == nil {
		return &ValidationError{Field: "request", Message: "at least one field must be provided for update"}
	}

	// Validate name if provided
	if req.Name != nil {
		if err := validateFunctionName(*req.Name); err != nil {
			return err
		}
	}

	// Validate description if provided
	if req.Description != nil {
		if err := validateDescription(*req.Description); err != nil {
			return err
		}
	}

	// Validate code if provided
	if req.Code != nil {
		if err := validateCode(*req.Code); err != nil {
			return err
		}
	}

	// Validate retention_days if provided
	if req.RetentionDays != nil {
		if err := validateRetentionDays(*req.RetentionDays); err != nil {
			return err
		}
	}

	// Validate cron_schedule if provided
	if req.CronSchedule != nil {
		if err := validateCronSchedule(*req.CronSchedule); err != nil {
			return err
		}
	}

	// Validate cron_status if provided
	if req.CronStatus != nil {
		if err := validateCronStatus(*req.CronStatus); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateEnvVarsRequest validates an UpdateEnvVarsRequest
func ValidateUpdateEnvVarsRequest(req *UpdateEnvVarsRequest) error {
	if req == nil {
		return &ValidationError{Field: "request", Message: "request cannot be nil"}
	}

	if req.EnvVars == nil {
		return &ValidationError{Field: "env_vars", Message: "env_vars cannot be nil"}
	}

	if len(req.EnvVars) > MaxEnvVars {
		return &ValidationError{
			Field:   "env_vars",
			Message: fmt.Sprintf("cannot have more than %d environment variables", MaxEnvVars),
		}
	}

	for key, value := range req.EnvVars {
		if err := validateEnvVarKey(key); err != nil {
			return err
		}
		if err := validateEnvVarValue(value); err != nil {
			return err
		}
	}

	return nil
}

// validateFunctionName validates a function name
func validateFunctionName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return &ValidationError{Field: "name", Message: "name cannot be empty"}
	}
	if len(trimmed) > MaxFunctionNameLength {
		return &ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("name cannot be longer than %d characters", MaxFunctionNameLength),
		}
	}
	return nil
}

// validateDescription validates a function description
func validateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return &ValidationError{
			Field:   "description",
			Message: fmt.Sprintf("description cannot be longer than %d characters", MaxDescriptionLength),
		}
	}
	return nil
}

// validateCode validates function code
func validateCode(code string) error {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return &ValidationError{Field: "code", Message: "code cannot be empty"}
	}
	if len(code) > MaxCodeLength {
		return &ValidationError{
			Field:   "code",
			Message: fmt.Sprintf("code cannot be longer than %d bytes", MaxCodeLength),
		}
	}
	return nil
}

// validateEnvVarKey validates an environment variable key
func validateEnvVarKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return &ValidationError{Field: "env_var_key", Message: "environment variable key cannot be empty"}
	}
	if len(key) > MaxEnvVarKeyLength {
		return &ValidationError{
			Field:   "env_var_key",
			Message: fmt.Sprintf("environment variable key cannot be longer than %d characters", MaxEnvVarKeyLength),
		}
	}
	// Additional validation: keys should only contain alphanumeric and underscores
	if !isValidEnvVarKey(key) {
		return &ValidationError{
			Field:   "env_var_key",
			Message: "environment variable key can only contain letters, numbers, and underscores",
		}
	}
	return nil
}

// validateEnvVarValue validates an environment variable value
func validateEnvVarValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return &ValidationError{Field: "env_var_value", Message: "environment variable value cannot be empty"}
	}
	if len(value) > MaxEnvVarValueLength {
		return &ValidationError{
			Field:   "env_var_value",
			Message: fmt.Sprintf("environment variable value cannot be longer than %d characters", MaxEnvVarValueLength),
		}
	}
	return nil
}

// isValidEnvVarKey checks if a string is a valid environment variable key
func isValidEnvVarKey(key string) bool {
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

// validateRetentionDays validates retention days value
func validateRetentionDays(days int) error {
	// Check if the value is in the allowed list
	if slices.Contains(AllowedRetentionDays, days) {
		return nil
	}
	return &ValidationError{
		Field:   "retention_days",
		Message: fmt.Sprintf("retention_days must be one of: %v", AllowedRetentionDays),
	}
}

// validateCronSchedule validates a cron expression
func validateCronSchedule(schedule string) error {
	// Empty schedule is allowed (to clear the schedule)
	if schedule == "" {
		return nil
	}

	// Parse the cron expression using the robfig/cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(schedule)
	if err != nil {
		return &ValidationError{
			Field:   "cron_schedule",
			Message: fmt.Sprintf("invalid cron expression: %v", err),
		}
	}
	return nil
}

// validateCronStatus validates a cron status value
func validateCronStatus(status string) error {
	if slices.Contains(AllowedCronStatuses, status) {
		return nil
	}
	return &ValidationError{
		Field:   "cron_status",
		Message: fmt.Sprintf("cron_status must be one of: %v", AllowedCronStatuses),
	}
}

// ValidateCronSchedule is a public function to validate cron expressions
// Used by the scheduler and handlers
func ValidateCronSchedule(schedule string) error {
	return validateCronSchedule(schedule)
}
