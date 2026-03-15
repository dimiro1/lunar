package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dimiro1/lunar/internal/token"
)

type Config struct {
	Port             string
	DataDir          string
	ExecutionTimeout time.Duration
	APIKey           string
	BaseURL          string
}

func loadPort(getenv func(string) string) string {
	port := getenv("PORT")
	if port == "" {
		port = "3000"
	}
	return port
}

func loadDataDir(getenv func(string) string, dataDir string) string {
	if dataDir == "" {
		dataDir = getenv("DATA_DIR")
		if dataDir == "" {
			dataDir = "./data"
		}
	}
	return dataDir
}

func loadTimeout(getenv func(string) string) time.Duration {
	timeoutStr := getenv("EXECUTION_TIMEOUT")
	timeout := 5 * time.Minute
	if timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(seconds) * time.Second
		}
	}
	return timeout
}

func generateAPIKey() (string, error) {
	return token.Generate()
}

func loadAPIKey(getenv func(string) string, dataDir string) (string, error) {
	// First, check environment variable
	apiKey := getenv("API_KEY")
	if apiKey != "" {
		return apiKey, nil
	}

	// Try to read from file
	apiKeyPath := filepath.Join(dataDir, "api_key.txt")
	keyBytes, err := os.ReadFile(apiKeyPath)
	if err == nil {
		return string(keyBytes), nil
	}

	// If file doesn't exist, generate new key
	if !os.IsNotExist(err) {
		return "", err
	}

	// Generate new API key
	apiKey, err = generateAPIKey()
	if err != nil {
		return "", err
	}

	// Save to file
	if err := os.WriteFile(apiKeyPath, []byte(apiKey), 0o600); err != nil {
		return "", err
	}

	slog.Info("Generated new API key", "key", apiKey, "file", apiKeyPath)
	return apiKey, nil
}

func loadBaseURL(getenv func(string) string, port string) string {
	baseURL := getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}
	return baseURL
}

func initDataDir(getenv func(string) string) (string, error) {
	dataDir := getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	return dataDir, nil
}

func loadConfig(getenv func(string) string, dataDir string) (Config, error) {
	port := loadPort(getenv)
	dataDir = loadDataDir(getenv, dataDir)
	timeout := loadTimeout(getenv)
	baseURL := loadBaseURL(getenv, port)

	apiKey, err := loadAPIKey(getenv, dataDir)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:             port,
		DataDir:          dataDir,
		ExecutionTimeout: timeout,
		APIKey:           apiKey,
		BaseURL:          baseURL,
	}, nil
}
