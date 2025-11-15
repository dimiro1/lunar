package main

import (
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	getenv := func(key string) string {
		return ""
	}

	config := loadConfig(getenv)

	if config.Port != "3000" {
		t.Errorf("expected default port 3000, got %s", config.Port)
	}

	if config.DataDir != "./data" {
		t.Errorf("expected default data dir ./data, got %s", config.DataDir)
	}

	if config.ExecutionTimeout != 5*time.Minute {
		t.Errorf("expected default timeout 5m, got %v", config.ExecutionTimeout)
	}
}

func TestLoadConfig_FromEnv(t *testing.T) {
	env := map[string]string{
		"PORT":              "8080",
		"DATA_DIR":          "/var/lib/faas",
		"EXECUTION_TIMEOUT": "60",
	}

	getenv := func(key string) string {
		return env[key]
	}

	config := loadConfig(getenv)

	if config.Port != "8080" {
		t.Errorf("expected port 8080, got %s", config.Port)
	}

	if config.DataDir != "/var/lib/faas" {
		t.Errorf("expected data dir /var/lib/faas, got %s", config.DataDir)
	}

	if config.ExecutionTimeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", config.ExecutionTimeout)
	}
}

func TestLoadConfig_InvalidTimeout(t *testing.T) {
	getenv := func(key string) string {
		if key == "EXECUTION_TIMEOUT" {
			return "invalid"
		}
		return ""
	}

	config := loadConfig(getenv)

	if config.ExecutionTimeout != 5*time.Minute {
		t.Errorf("expected default timeout 5m for invalid input, got %v", config.ExecutionTimeout)
	}
}

func TestLoadConfig_PartialEnv(t *testing.T) {
	getenv := func(key string) string {
		if key == "PORT" {
			return "9000"
		}
		return ""
	}

	config := loadConfig(getenv)

	if config.Port != "9000" {
		t.Errorf("expected port 9000, got %s", config.Port)
	}

	if config.DataDir != "./data" {
		t.Errorf("expected default data dir ./data, got %s", config.DataDir)
	}

	if config.ExecutionTimeout != 5*time.Minute {
		t.Errorf("expected default timeout 5m, got %v", config.ExecutionTimeout)
	}
}
