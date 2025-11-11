# Build variables
VERSION ?= $(shell git describe --tags --always --dirty --match="v*" 2> /dev/null || echo "dev")
BUILD_DATE = $(shell date -u '+%Y-%m-%d')
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Go variables
GO = go
GOBIN = $(shell $(GO) env GOPATH)/bin

# Docker variables
IMAGE = razorpay-mcp-server
TAG ?= latest

# Default target
all: verify fmt test lint build

# Build docker image
build:
	docker build \
		--build-arg VERSION="$(VERSION)" \
		-t $(IMAGE):$(TAG) \
		.

# Run docker container
run:
	docker run -it --rm \
		-e RAZORPAY_KEY_ID=your_key_id \
		-e RAZORPAY_KEY_SECRET=your_key_secret \
		$(IMAGE):$(TAG)

# Run the application
local-run:
	$(GO) run ./cmd/razorpay-mcp-server

local-build:
	$(GO) build $(LDFLAGS) -v -o razorpay-mcp-server ./cmd/razorpay-mcp-server

# Verify dependencies
verify:
	$(GO) mod verify
	$(GO) mod download

# Format code
fmt:
	$(GO) fmt ./...
	$(GO) mod tidy

# Run tests
test:
	$(GO) test -race ./...

# Run tests with coverage
test-coverage:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./pkg/...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Install golangci-lint
install-lint:
	@if [ ! -f ./bin/golangci-lint ]; then \
		echo "Building golangci-lint from source with current Go version..."; \
		mkdir -p bin; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		if [ -f $$($(GO) env GOPATH)/bin/golangci-lint ]; then \
			cp $$($(GO) env GOPATH)/bin/golangci-lint bin/; \
			echo "golangci-lint installed to bin/golangci-lint"; \
		else \
			echo "Warning: Failed to install golangci-lint. Using system version if available."; \
		fi \
	fi

# Run linter
lint: install-lint
	@if [ -f ./bin/golangci-lint ]; then \
		./bin/golangci-lint run --out-format=colored-line-number --timeout=3m || true; \
	elif command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run --out-format=colored-line-number --timeout=3m || true; \
	else \
		echo "Warning: golangci-lint not found. Skipping lint check."; \
		echo "Please install golangci-lint manually: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  all            - Run verify, fmt, test, lint, and build (default)"
	@echo "  build          - Build Docker image"
	@echo "  run            - Run Docker container"
	@echo "  local-build    - Build the application"
	@echo "  local-run      - Run the application"
	@echo "  verify         - Verify dependencies"
	@echo "  fmt            - Format code"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message" 