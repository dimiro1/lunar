package api

import (
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/store"
)

func TestValidateCreateFunctionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateFunctionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &CreateFunctionRequest{
				Name: "test-function",
				Code: "function handler() end",
			},
			wantErr: false,
		},
		{
			name: "valid request with description",
			req: &CreateFunctionRequest{
				Name:        "test-function",
				Description: strPtr("A test function"),
				Code:        "function handler() end",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name: "empty name",
			req: &CreateFunctionRequest{
				Name: "",
				Code: "function handler() end",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "whitespace only name",
			req: &CreateFunctionRequest{
				Name: "   ",
				Code: "function handler() end",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "name too long",
			req: &CreateFunctionRequest{
				Name: strings.Repeat("a", MaxFunctionNameLength+1),
				Code: "function handler() end",
			},
			wantErr: true,
			errMsg:  "name cannot be longer",
		},
		{
			name: "empty code",
			req: &CreateFunctionRequest{
				Name: "test-function",
				Code: "",
			},
			wantErr: true,
			errMsg:  "code cannot be empty",
		},
		{
			name: "whitespace only code",
			req: &CreateFunctionRequest{
				Name: "test-function",
				Code: "   \n  \t  ",
			},
			wantErr: true,
			errMsg:  "code cannot be empty",
		},
		{
			name: "code too long",
			req: &CreateFunctionRequest{
				Name: "test-function",
				Code: strings.Repeat("a", MaxCodeLength+1),
			},
			wantErr: true,
			errMsg:  "code cannot be longer",
		},
		{
			name: "description too long",
			req: &CreateFunctionRequest{
				Name:        "test-function",
				Description: strPtr(strings.Repeat("a", MaxDescriptionLength+1)),
				Code:        "function handler() end",
			},
			wantErr: true,
			errMsg:  "description cannot be longer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateCreateFunctionRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUpdateFunctionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *store.UpdateFunctionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with name",
			req: &store.UpdateFunctionRequest{
				Name: strPtr("new-name"),
			},
			wantErr: false,
		},
		{
			name: "valid request with code",
			req: &store.UpdateFunctionRequest{
				Code: strPtr("function handler() end"),
			},
			wantErr: false,
		},
		{
			name: "valid request with all fields",
			req: &store.UpdateFunctionRequest{
				Name:        strPtr("new-name"),
				Description: strPtr("new description"),
				Code:        strPtr("function handler() end"),
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name:    "empty request",
			req:     &store.UpdateFunctionRequest{},
			wantErr: true,
			errMsg:  "at least one field must be provided",
		},
		{
			name: "invalid name",
			req: &store.UpdateFunctionRequest{
				Name: strPtr(""),
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "invalid code",
			req: &store.UpdateFunctionRequest{
				Code: strPtr(""),
			},
			wantErr: true,
			errMsg:  "code cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUpdateEnvVarsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateEnvVarsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"API_KEY": "secret",
					"PORT":    "3000",
				},
			},
			wantErr: false,
		},
		{
			name: "valid empty env vars",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{},
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name: "nil env vars",
			req: &UpdateEnvVarsRequest{
				EnvVars: nil,
			},
			wantErr: true,
			errMsg:  "env_vars cannot be nil",
		},
		{
			name: "invalid env var key - empty",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"": "value",
				},
			},
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name: "invalid env var key - special chars",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"API-KEY": "value",
				},
			},
			wantErr: true,
			errMsg:  "can only contain letters, numbers, and underscores",
		},
		{
			name: "invalid env var key - too long",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					strings.Repeat("A", MaxEnvVarKeyLength+1): "value",
				},
			},
			wantErr: true,
			errMsg:  "key cannot be longer",
		},
		{
			name: "invalid env var value - empty",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"KEY": "",
				},
			},
			wantErr: true,
			errMsg:  "value cannot be empty",
		},
		{
			name: "invalid env var value - whitespace only",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"KEY": "   ",
				},
			},
			wantErr: true,
			errMsg:  "value cannot be empty",
		},
		{
			name: "invalid env var value - too long",
			req: &UpdateEnvVarsRequest{
				EnvVars: map[string]string{
					"KEY": strings.Repeat("a", MaxEnvVarValueLength+1),
				},
			},
			wantErr: true,
			errMsg:  "value cannot be longer",
		},
		{
			name: "too many env vars",
			req: &UpdateEnvVarsRequest{
				EnvVars: func() map[string]string {
					m := make(map[string]string)
					for i := 0; i < MaxEnvVars+1; i++ {
						m[strings.Repeat("A", i%10+1)+string(rune(i))] = "value"
					}
					return m
				}(),
			},
			wantErr: true,
			errMsg:  "cannot have more than",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateEnvVarsRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateEnvVarsRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateUpdateEnvVarsRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestIsValidEnvVarKey(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		{"API_KEY", true},
		{"PORT", true},
		{"DB_HOST_1", true},
		{"_PRIVATE", true},
		{"snake_case_key", true},
		{"UPPERCASE_KEY", true},
		{"MixedCase123", true},
		{"", false},
		{"API-KEY", false},
		{"API.KEY", false},
		{"API KEY", false},
		{"API@KEY", false},
		{"123", true},
		{"_123", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isValidEnvVarKey(tt.key); got != tt.valid {
				t.Errorf("isValidEnvVarKey(%q) = %v, want %v", tt.key, got, tt.valid)
			}
		})
	}
}

// Helper function for creating string pointers
func strPtr(s string) *string {
	return &s
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}

func TestValidateRetentionDays(t *testing.T) {
	tests := []struct {
		name    string
		days    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid 7 days",
			days:    7,
			wantErr: false,
		},
		{
			name:    "valid 15 days",
			days:    15,
			wantErr: false,
		},
		{
			name:    "valid 30 days",
			days:    30,
			wantErr: false,
		},
		{
			name:    "valid 365 days (1 year)",
			days:    365,
			wantErr: false,
		},
		{
			name:    "invalid 1 day",
			days:    1,
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "invalid 5 days",
			days:    5,
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "invalid 60 days",
			days:    60,
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "invalid 500 days",
			days:    500,
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "invalid 0 days",
			days:    0,
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:    "invalid negative days",
			days:    -1,
			wantErr: true,
			errMsg:  "must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRetentionDays(tt.days)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRetentionDays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateRetentionDays() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUpdateFunctionRequest_WithRetentionDays(t *testing.T) {
	tests := []struct {
		name    string
		req     *store.UpdateFunctionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid retention days 7",
			req: &store.UpdateFunctionRequest{
				RetentionDays: intPtr(7),
			},
			wantErr: false,
		},
		{
			name: "valid retention days 30",
			req: &store.UpdateFunctionRequest{
				RetentionDays: intPtr(30),
			},
			wantErr: false,
		},
		{
			name: "invalid retention days 5",
			req: &store.UpdateFunctionRequest{
				RetentionDays: intPtr(5),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "invalid retention days 100",
			req: &store.UpdateFunctionRequest{
				RetentionDays: intPtr(100),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "combined update with valid retention days",
			req: &store.UpdateFunctionRequest{
				Name:          strPtr("new-name"),
				RetentionDays: intPtr(15),
			},
			wantErr: false,
		},
		{
			name: "combined update with invalid retention days",
			req: &store.UpdateFunctionRequest{
				Name:          strPtr("new-name"),
				RetentionDays: intPtr(20),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUpdateFunctionRequest_WithCronSchedule(t *testing.T) {
	tests := []struct {
		name    string
		req     *store.UpdateFunctionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cron schedule - every 5 minutes",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("*/5 * * * *"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule - every hour",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("0 * * * *"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule - every day at midnight",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("0 0 * * *"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule - weekdays at 9am",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("0 9 * * 1-5"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule - first day of month",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("0 0 1 * *"),
			},
			wantErr: false,
		},
		{
			name: "valid empty cron schedule (to clear)",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr(""),
			},
			wantErr: false,
		},
		{
			name: "invalid cron schedule - too few fields",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("* * *"),
			},
			wantErr: true,
			errMsg:  "invalid cron expression",
		},
		{
			name: "invalid cron schedule - too many fields",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("* * * * * *"),
			},
			wantErr: true,
			errMsg:  "invalid cron expression",
		},
		{
			name: "invalid cron schedule - bad minute value",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("60 * * * *"),
			},
			wantErr: true,
			errMsg:  "invalid cron expression",
		},
		{
			name: "invalid cron schedule - bad hour value",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("* 25 * * *"),
			},
			wantErr: true,
			errMsg:  "invalid cron expression",
		},
		{
			name: "invalid cron schedule - invalid syntax",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("not a cron"),
			},
			wantErr: true,
			errMsg:  "invalid cron expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUpdateFunctionRequest_WithSaveResponse(t *testing.T) {
	tests := []struct {
		name    string
		req     *store.UpdateFunctionRequest
		wantErr bool
	}{
		{
			name: "valid save_response true",
			req: &store.UpdateFunctionRequest{
				SaveResponse: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "valid save_response false",
			req: &store.UpdateFunctionRequest{
				SaveResponse: boolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "combined update with save_response",
			req: &store.UpdateFunctionRequest{
				Name:         strPtr("new-name"),
				SaveResponse: boolPtr(true),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func TestValidateUpdateFunctionRequest_WithCronStatus(t *testing.T) {
	tests := []struct {
		name    string
		req     *store.UpdateFunctionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cron status - active",
			req: &store.UpdateFunctionRequest{
				CronStatus: strPtr("active"),
			},
			wantErr: false,
		},
		{
			name: "valid cron status - paused",
			req: &store.UpdateFunctionRequest{
				CronStatus: strPtr("paused"),
			},
			wantErr: false,
		},
		{
			name: "invalid cron status - stopped",
			req: &store.UpdateFunctionRequest{
				CronStatus: strPtr("stopped"),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "invalid cron status - enabled",
			req: &store.UpdateFunctionRequest{
				CronStatus: strPtr("enabled"),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "invalid cron status - empty",
			req: &store.UpdateFunctionRequest{
				CronStatus: strPtr(""),
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "valid combined cron schedule and status",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("*/5 * * * *"),
				CronStatus:   strPtr("active"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule with paused status",
			req: &store.UpdateFunctionRequest{
				CronSchedule: strPtr("0 * * * *"),
				CronStatus:   strPtr("paused"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateFunctionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateUpdateFunctionRequest() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}
