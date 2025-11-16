# DevSmith Platform Makefile
# Build all services with proper version injection

# Version information (extracted from git)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_NUMBER := $(shell echo $${BUILD_NUMBER:-0})

# Go build flags with version injection
LDFLAGS := -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.Version=$(VERSION) \
           -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.CommitHash=$(COMMIT) \
           -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.BuildTime=$(BUILD_TIME) \
           -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.BuildNumber=$(BUILD_NUMBER)

# Build directories
BIN_DIR := bin
BUILD_DIR := build

# Targets
.PHONY: all build-logs build-portal build-analytics build-review build test clean version help templ fmt lint

all: build

build: build-logs build-portal build-analytics build-review

## Build Commands

build-logs: ## Build logs service with version injection
	@echo "Building logs service..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/logs ./cmd/logs
	@echo "✓ Logs service built: $(BIN_DIR)/logs"

build-portal: ## Build portal service with version injection
	@echo "Building portal service..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/portal ./cmd/portal
	@echo "✓ Portal service built: $(BIN_DIR)/portal"

build-analytics: ## Build analytics service with version injection
	@echo "Building analytics service..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/analytics ./cmd/analytics
	@echo "✓ Analytics service built: $(BIN_DIR)/analytics"

build-review: ## Build review service with version injection
	@echo "Building review service..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/review ./cmd/review
	@echo "✓ Review service built: $(BIN_DIR)/review"

## Test Commands

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-version: ## Test version package specifically
	@echo "Testing version package..."
	go test -v ./internal/version/...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## Utility Commands

version: ## Display version information
	@echo "DevSmith Platform Build Information"
	@echo "===================================="
	@echo "Version:      $(VERSION)"
	@echo "Commit:       $(COMMIT)"
	@echo "Build Time:   $(BUILD_TIME)"
	@echo "Build Number: $(BUILD_NUMBER)"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

templ: ## Regenerate Templ templates
	@echo "Regenerating Templ templates..."
	templ generate
	@echo "✓ Templates regenerated"

fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

docker-build-logs: ## Build logs Docker image with version
	@echo "Building logs Docker image..."
	docker build -f Dockerfile.logs \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t devsmith-logs:$(VERSION) \
		-t devsmith-logs:latest \
		.
	@echo "✓ Docker image built: devsmith-logs:$(VERSION)"

docker-build-portal: ## Build portal Docker image with version
	@echo "Building portal Docker image..."
	docker build -f Dockerfile.portal \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t devsmith-portal:$(VERSION) \
		-t devsmith-portal:latest \
		.
	@echo "✓ Docker image built: devsmith-portal:$(VERSION)"

help: ## Show this help message
	@echo "DevSmith Platform - Available Make Targets"
	@echo "=========================================="
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
