.PHONY: build test lint clean run help

BINARY_NAME=faas-go
BUILD_DIR=build

help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run all tests"
	@echo "  lint     - Run golangci-lint"
	@echo "  clean    - Remove build artifacts"
	@echo "  run      - Build and run the application"
	@echo "  all      - Run lint, test, and build"

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

test:
	@echo "Running tests..."
	@go test ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

run: build
	@echo "Running application..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

all: lint test build
	@echo "All checks passed!"
