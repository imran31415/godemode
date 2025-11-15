.PHONY: help build run test clean install-deps check-tinygo

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

check-tinygo: ## Check if TinyGo is installed
	@which tinygo > /dev/null || (echo "Error: TinyGo is not installed. Install with: brew install tinygo" && exit 1)
	@echo "TinyGo is installed: $$(tinygo version)"

install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

build: check-tinygo ## Build the godemode binary
	go build -o godemode ./cmd/godemode

run: build ## Build and run with example
	./godemode --help

test: ## Run tests
	go test -v ./pkg/... ./cmd/... ./internal/...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./pkg/... ./cmd/... ./internal/...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f godemode
	rm -f coverage.out coverage.html
	go clean

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./pkg/... ./cmd/... ./internal/...

lint: ## Run golangci-lint (if installed)
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed, skipping"

all: clean install-deps fmt vet build test ## Clean, build, and test everything
