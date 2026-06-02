package logging

import (
	"log/slog"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		want    slog.Level
		wantErr bool
	}{
		{"unset defaults to info", map[string]string{}, slog.LevelInfo, false},
		{"debug lowercase", map[string]string{"LOG_LEVEL": "debug"}, slog.LevelDebug, false},
		{"debug uppercase", map[string]string{"LOG_LEVEL": "DEBUG"}, slog.LevelDebug, false},
		{"warn", map[string]string{"LOG_LEVEL": "warn"}, slog.LevelWarn, false},
		{"error uppercase", map[string]string{"LOG_LEVEL": "ERROR"}, slog.LevelError, false},
		{"invalid falls back to info", map[string]string{"LOG_LEVEL": "nonsense"}, slog.LevelInfo, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := load(tc.env)
			if (err != nil) != tc.wantErr {
				t.Fatalf("load(%v) error = %v, wantErr %v", tc.env, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("load(%v) = %v, want %v", tc.env, got, tc.want)
			}
		})
	}
}
