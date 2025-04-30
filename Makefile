# Build variables
VERSION ?= $(shell git describe --tags --always --dirty --match="v*" 2> /dev/null || echo "dev")
BUILD_DATE = $(shell date -u '+%Y-%m-%d')
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

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
	docker build -t $(IMAGE):$(TAG) .

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
	$(GO) build -v ./cmd/razorpay-mcp-server

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
	@LINT_VERSION=1.64.8; \
	if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "Installing golangci-lint v$$LINT_VERSION..."; \
		curl -fsSL https://github.com/golangci/golangci-lint/releases/download/v$$LINT_VERSION/golangci-lint-$$LINT_VERSION-$$($(GO) env GOOS)-$$($(GO) env GOARCH).tar.gz | \
			tar xz --strip-components 1 --wildcards \*/golangci-lint; \
		mkdir -p bin && mv golangci-lint bin/; \
		echo "golangci-lint installed to bin/golangci-lint"; \
	fi

# Run linter
lint: install-lint
	@if [ -f ./bin/golangci-lint ]; then \
		./bin/golangci-lint run --out-format=colored-line-number --timeout=3m; \
	else \
		golangci-lint run --out-format=colored-line-number --timeout=3m; \
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