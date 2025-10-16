.PHONY: help build build-all install install-local test lint fmt vet clean run deps tidy check test-fixtures test-fixtures-clean

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := sortpics
CMD_PATH := ./cmd/sortpics
BUILD_DIR := ./bin
DIST_DIR := ./dist
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w"
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
INSTALL_DIR := $(HOME)/.local/bin

# Platform-specific variables
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary for the current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all supported platforms
build-all:
	@echo "Building $(BINARY_NAME) $(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)
	# macOS
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	GOOS=windows GOARCH=arm64 $(GO) build $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe $(CMD_PATH)
	@echo "Binaries built in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

## install: Install the binary to ~/.local/bin (recommended)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""
	@echo ""
	@echo "Add this to your ~/.bashrc, ~/.zshrc, or ~/.profile to make it permanent"

## install-global: Install the binary to GOPATH/bin (requires Go in PATH)
install-global:
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	$(GO) install $(GOFLAGS) $(LDFLAGS) -ldflags "-X main.version=$(VERSION)" $(CMD_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -cover ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	$(GO) test $(GOFLAGS) -race -cover -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## lint: Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	$(GO) mod tidy
	@echo "Modules tidied"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

## check: Run fmt, vet, and test
check: fmt vet test
	@echo "All checks passed"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Build and run the binary (use ARGS for arguments)
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## run-dev: Run directly with go run (use ARGS for arguments)
run-dev:
	@echo "Running with go run..."
	$(GO) run $(CMD_PATH) $(ARGS)

## test-fixtures: Generate test fixtures for integration tests
test-fixtures:
	@echo "Generating test fixtures..."
	@which exiftool > /dev/null || (echo "exiftool not installed. Run: brew install exiftool" && exit 1)
	@cd test/testdata && go run generate_fixtures.go
	@echo "Test fixtures generated in test/testdata/"

## test-fixtures-clean: Remove generated test fixtures
test-fixtures-clean:
	@echo "Removing test fixtures..."
	@rm -rf test/testdata/basic test/testdata/mixed test/testdata/no_exif
	@rm -rf test/testdata/special_makes test/testdata/collision test/testdata/video
	@rm -f test/testdata/manifest.json
	@echo "Test fixtures removed"
