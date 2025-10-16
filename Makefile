.PHONY: help build test lint fmt vet clean install run deps tidy check test-fixtures test-fixtures-clean

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := sortpics
CMD_PATH := ./cmd/sortpics
BUILD_DIR := ./bin
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w"

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) $(LDFLAGS) $(CMD_PATH)
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
	@rm -rf $(BUILD_DIR)
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
