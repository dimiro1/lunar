.PHONY: build test lint clean run help dev install-tools fmt-frontend test-frontend test-e2e test-all vendor-js

BINARY_NAME=lunar
BUILD_DIR=build

# Frontend vendor dependency versions
MITHRIL_VERSION=2.3.7
MONACO_VERSION=0.54.0
HIGHLIGHTJS_VERSION=11.11.1
JASMINE_VERSION=5.4.0

help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run Go unit tests"
	@echo "  test-frontend - Open Jasmine frontend tests in browser"
	@echo "  test-e2e      - Run E2E tests with headless Chrome"
	@echo "  test-all      - Run all tests (unit + e2e)"
	@echo "  lint          - Run golangci-lint"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  all           - Run lint, test, and build"
	@echo "  fmt-frontend  - Format frontend JS files with deno fmt"
	@echo "  vendor-js     - Download/update vendored JS dependencies"

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

test:
	@echo "Running tests..."
	@go test $$(go list ./... | grep -v /e2e)

lint:
	@echo "Running linter..."
	@golangci-lint run

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

run:
	@echo "Running application..."
	@go run ./cmd

all: lint test build
	@echo "All checks passed!"

dev:
	@echo "Starting development mode with air..."
	@air

install-tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest

fmt-frontend:
	@echo "Formatting frontend..."
	@deno fmt --ignore=frontend/vendor frontend/

test-frontend:
	@go run ./cmd/testserver

test-e2e:
	@echo "Running E2E tests with headless Chrome..."
	@go test -v ./e2e/...

test-all: test test-e2e
	@echo "All Go tests passed!"
	@echo "Run 'make test-frontend' to open browser tests manually"

vendor-js:
	@echo "Downloading vendored JS dependencies..."
	@mkdir -p frontend/vendor/{mithril,monaco-editor,highlight.js/styles,highlight.js/languages,jasmine}
	@echo "Downloading Mithril.js $(MITHRIL_VERSION)..."
	@curl -sL -o frontend/vendor/mithril/mithril.min.js \
		"https://unpkg.com/mithril@$(MITHRIL_VERSION)/mithril.min.js"
	@echo "Downloading Monaco Editor $(MONACO_VERSION)..."
	@curl -sL "https://registry.npmjs.org/monaco-editor/-/monaco-editor-$(MONACO_VERSION).tgz" | \
		tar -xz -C frontend/vendor/monaco-editor --strip-components=1 package/min
	@echo "Downloading Highlight.js $(HIGHLIGHTJS_VERSION)..."
	@curl -sL -o frontend/vendor/highlight.js/highlight.min.js \
		"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/$(HIGHLIGHTJS_VERSION)/highlight.min.js"
	@curl -sL -o frontend/vendor/highlight.js/styles/github-dark.min.css \
		"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/$(HIGHLIGHTJS_VERSION)/styles/github-dark.min.css"
	@for lang in lua bash javascript python go json; do \
		curl -sL -o "frontend/vendor/highlight.js/languages/$${lang}.min.js" \
			"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/$(HIGHLIGHTJS_VERSION)/languages/$${lang}.min.js"; \
	done
	@echo "Downloading Jasmine $(JASMINE_VERSION)..."
	@for file in jasmine.css jasmine.js jasmine-html.js boot0.js boot1.js; do \
		curl -sL -o "frontend/vendor/jasmine/$${file}" \
			"https://unpkg.com/jasmine-core@$(JASMINE_VERSION)/lib/jasmine-core/$${file}"; \
	done
	@echo "Done! Vendored JS dependencies updated."
