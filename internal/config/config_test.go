package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse_Defaults(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "3000" {
		t.Errorf("expected default port 3000, got %s", cfg.Port)
	}
	if cfg.DataDir != tmpDir {
		t.Errorf("expected data dir %s, got %s", tmpDir, cfg.DataDir)
	}
	if cfg.ExecutionTimeout != 5*time.Minute {
		t.Errorf("expected default timeout 5m, got %v", cfg.ExecutionTimeout)
	}
	if cfg.BaseURL != "http://localhost:3000" {
		t.Errorf("expected base URL http://localhost:3000, got %s", cfg.BaseURL)
	}
	if len(cfg.APIKey) != 64 {
		t.Errorf("expected 64-character generated API key, got %d characters", len(cfg.APIKey))
	}
}

func TestParse_FromEnv(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{
		"PORT":              "8080",
		"DATA_DIR":          tmpDir,
		"EXECUTION_TIMEOUT": "60",
		"API_KEY":           "custom-api-key",
		"BASE_URL":          "https://lunar.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.DataDir != tmpDir {
		t.Errorf("expected data dir %s, got %s", tmpDir, cfg.DataDir)
	}
	if cfg.ExecutionTimeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.ExecutionTimeout)
	}
	if cfg.APIKey != "custom-api-key" {
		t.Errorf("expected API key 'custom-api-key', got %s", cfg.APIKey)
	}
	if cfg.BaseURL != "https://lunar.example.com" {
		t.Errorf("expected base URL 'https://lunar.example.com', got %s", cfg.BaseURL)
	}
}

func TestParse_ExecutionTimeoutSeconds(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir, "EXECUTION_TIMEOUT": "120"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ExecutionTimeout != 120*time.Second {
		t.Errorf("expected timeout 120s, got %v", cfg.ExecutionTimeout)
	}
}

func TestParse_ExecutionTimeoutInvalidFallsBackToDefault(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir, "EXECUTION_TIMEOUT": "not-a-number"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ExecutionTimeout != 5*time.Minute {
		t.Errorf("expected default timeout 5m for invalid input, got %v", cfg.ExecutionTimeout)
	}
}

func TestParse_BaseURLDefaultsToLocalhostWithPort(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir, "PORT": "4000"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BaseURL != "http://localhost:4000" {
		t.Errorf("expected base URL http://localhost:4000, got %s", cfg.BaseURL)
	}
}

func TestParse_APIKeyFromEnvDoesNotWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir, "API_KEY": "test-key-from-env"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "test-key-from-env" {
		t.Errorf("expected key 'test-key-from-env', got %s", cfg.APIKey)
	}

	keyPath := filepath.Join(tmpDir, apiKeyFile)
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Error("api_key.txt should not be created when API_KEY env var is set")
	}
}

func TestParse_APIKeyFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, apiKeyFile)

	expectedKey := "test-key-from-file"
	if err := os.WriteFile(keyPath, []byte(expectedKey), 0o600); err != nil {
		t.Fatalf("failed to create test key file: %v", err)
	}

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != expectedKey {
		t.Errorf("expected key '%s', got %s", expectedKey, cfg.APIKey)
	}
}

func TestParse_APIKeyGeneratedAndPersisted(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.APIKey) != 64 {
		t.Errorf("expected 64-character key, got %d characters", len(cfg.APIKey))
	}

	keyPath := filepath.Join(tmpDir, apiKeyFile)
	savedKey, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read saved key file: %v", err)
	}
	if string(savedKey) != cfg.APIKey {
		t.Error("saved key doesn't match returned key")
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected file permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestParse_APIKeyEnvTakesPrecedenceOverFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, apiKeyFile)

	if err := os.WriteFile(keyPath, []byte("file-key"), 0o600); err != nil {
		t.Fatalf("failed to create test key file: %v", err)
	}

	cfg, err := parse(map[string]string{"DATA_DIR": tmpDir, "API_KEY": "env-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "env-key" {
		t.Errorf("expected env key 'env-key', got %s", cfg.APIKey)
	}
}
