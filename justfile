# JustGoHTML Justfile - HTML5 parser build automation
# Run `just` or `just help` to see available commands

set shell := ["bash", "-euo", "pipefail", "-c"]

# Version from file (defaults to dev if not present)
VERSION := `cat ./version.txt 2>/dev/null || echo "dev"`

# Binary name
BINARY := "JustGoHTML"

# Linting configuration - relax some checks during initial development
_LINT_FLAGS := "--timeout 2m"

# Default recipe - show help
default:
    @just --list

# Show help information
help:
    @echo "JustGoHTML - HTML5 Parser Build Automation"
    @echo ""
    @echo "Available commands:"
    @just --list

#################################
# Development Setup
#################################

# Install all required formatters and development tools
setup-deps:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing required formatters and tools..."

    # Go-based formatters
    command -v gofumpt >/dev/null 2>&1 || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }
    command -v gci >/dev/null 2>&1 || { echo "Installing gci..."; go install github.com/daixiang0/gci@latest; }
    command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2; }

    # Node-based formatters
    command -v prettier >/dev/null 2>&1 || { echo "Installing prettier..."; npm install -g prettier || echo "prettier installation failed - npm not found"; }

    # Rust-based formatters
    command -v taplo >/dev/null 2>&1 || { echo "Installing taplo..."; cargo install taplo-cli --version "0.9.3" || echo "taplo installation failed - cargo not found"; }
    command -v treefmt >/dev/null 2>&1 || { echo "Installing treefmt..."; cargo install treefmt || echo "treefmt installation failed - cargo not found"; }

    # System tools
    command -v shfmt >/dev/null 2>&1 || echo "shfmt not found. Please install: https://github.com/mvdan/sh/releases"
    command -v shellcheck >/dev/null 2>&1 || echo "shellcheck not found. Please install: https://github.com/koalaman/shellcheck#installing"

    echo "Dependencies installation completed!"
    echo "Note: Ensure $(go env GOPATH)/bin is in your PATH"

# Setup development environment
setup:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Checking required tools..."
    command -v go >/dev/null || { echo "Error: go not found"; exit 1; }
    command -v golangci-lint >/dev/null || echo "Warning: golangci-lint not found (run 'just setup-deps')"
    command -v treefmt >/dev/null || echo "Warning: treefmt not found (run 'just setup-deps')"
    echo "Setup complete! Run 'just check' to verify everything works."

#################################
# Code Quality
#################################

# Format all code using treefmt
fmt:
    treefmt --allow-missing-formatter

# Run linter
lint *FLAGS="":
    golangci-lint run {{_LINT_FLAGS}} {{FLAGS}} ./...

# Fix linting issues automatically where possible
lint-fix:
    golangci-lint run {{_LINT_FLAGS}} --fix ./...

#################################
# Testing
#################################

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Run tests with coverage report
test-coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    go tool cover -func=coverage.out | tail -1
    echo "Coverage report saved to coverage.html"

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run html5lib-tests (verbose)
test-spec:
    go test -v -timeout 30m -run 'HTML5Lib' ./...

# Run html5lib-tests and fail on mismatches (tree construction is progress-mode by default)
test-spec-strict:
    JustGoHTML_HTML5LIB_STRICT=1 go test -v -timeout 30m -run 'HTML5Lib' ./...

#################################
# Validation & Checks
#################################

# Run all checks (formatting, tests, linting)
check: check-formatted test lint

# Check that code is properly formatted
check-formatted:
    treefmt --fail-on-change --allow-missing-formatter

# Comprehensive check with coverage (for CI/CD)
check-ci: check test-coverage
    @echo "All CI checks passed!"

#################################
# Build Commands
#################################

# Build the CLI binary
build:
    go build -ldflags '-s -w -X main.version={{VERSION}}' -o {{BINARY}} ./cmd/JustGoHTML

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for Linux
build-linux:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -X main.version={{VERSION}}' -o {{BINARY}}-linux-amd64 ./cmd/JustGoHTML

# Build for macOS (Intel)
build-darwin:
    GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w -X main.version={{VERSION}}' -o {{BINARY}}-darwin-amd64 ./cmd/JustGoHTML

# Build for macOS (Apple Silicon)
build-darwin-arm:
    GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w -X main.version={{VERSION}}' -o {{BINARY}}-darwin-arm64 ./cmd/JustGoHTML

# Build for Windows
build-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags '-s -w -X main.version={{VERSION}}' -o {{BINARY}}.exe ./cmd/JustGoHTML

# Install locally
install:
    go install -ldflags '-s -w -X main.version={{VERSION}}' ./cmd/JustGoHTML

#################################
# WebAssembly Build
#################################

# Build WebAssembly module
build-wasm:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building WebAssembly module..."
    mkdir -p playground
    GOOS=js GOARCH=wasm go build -ldflags '-s -w' -o playground/justgohtml.wasm ./cmd/wasm
    # Copy wasm_exec.js from Go installation
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" playground/
    echo "WASM build complete: playground/justgohtml.wasm"
    ls -lh playground/justgohtml.wasm

# Build optimized WebAssembly module (smaller, using tinygo if available)
build-wasm-tiny:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v tinygo >/dev/null 2>&1; then
        echo "Building optimized WebAssembly module with TinyGo..."
        mkdir -p playground
        tinygo build -o playground/justgohtml.wasm -target wasm ./cmd/wasm
        cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" playground/
        echo "TinyGo WASM build complete: playground/justgohtml.wasm"
        ls -lh playground/justgohtml.wasm
    else
        echo "TinyGo not found, using standard Go build"
        just build-wasm
    fi

# Serve playground locally for development
serve-playground:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -f playground/justgohtml.wasm ]; then
        echo "Building WASM first..."
        just build-wasm
    fi
    echo "Starting playground server at http://localhost:8080"
    cd playground && python3 -m http.server 8080

#################################
# Development
#################################

# Run the CLI (for testing during development)
run *ARGS="":
    go run ./cmd/JustGoHTML {{ARGS}}

# Generate code (entities, constants)
gen:
    @echo "TODO: Add code generation for HTML5 entities"

# Tidy go modules
tidy:
    go mod tidy

#################################
# Utility Commands
#################################

# Clean build artifacts
clean:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Cleaning build artifacts..."
    rm -f {{BINARY}} {{BINARY}}-* {{BINARY}}.exe
    rm -f coverage.out coverage.html
    rm -f playground/justgohtml.wasm playground/wasm_exec.js
    echo "Clean completed!"

# Show project version and build info
version:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "JustGoHTML version: {{VERSION}}"
    echo "Go version: $(go version)"

# Watch for changes and run tests
watch:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v fswatch >/dev/null 2>&1; then
        echo "Watching for changes... (Ctrl+C to stop)"
        fswatch -o . --exclude=".git" --exclude="coverage.*" --exclude="{{BINARY}}*" -e ".*" -i "\\.go$" | while read; do
            echo "Changes detected, running tests..."
            just test || echo "Tests failed"
        done
    elif command -v inotifywait >/dev/null 2>&1; then
        echo "Watching for changes... (Ctrl+C to stop)"
        while inotifywait -r -e modify --include='\.go$' .; do
            echo "Changes detected, running tests..."
            just test || echo "Tests failed"
        done
    else
        echo "Neither fswatch nor inotifywait found. Install one of them for watch mode."
        exit 1
    fi

# Show lines of code statistics
loc:
    @echo "Go source files:"
    @find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
