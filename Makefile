.PHONY: help build run test clean install-deps check-tinygo docker-build docker-push setup-namespace k8s-deploy deploy

# Docker and Kubernetes variables
REGISTRY = registry.digitalocean.com/resourceloop
IMAGE_NAME = godemode/frontend
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
TIMESTAMP = $(shell date +%Y%m%d-%H%M%S)
IMAGE_TAG = $(GIT_COMMIT)-$(TIMESTAMP)
FULL_IMAGE = $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
LATEST_IMAGE = $(REGISTRY)/$(IMAGE_NAME):latest

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

# Docker targets
docker-build: ## Build frontend Docker image
	@echo "Building frontend Docker image: $(FULL_IMAGE)"
	cd frontend && docker build --platform linux/amd64 -f Dockerfile -t $(FULL_IMAGE) -t $(LATEST_IMAGE) .
	@echo "Built: $(FULL_IMAGE)"
	@echo "Tagged: $(LATEST_IMAGE)"

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing $(LATEST_IMAGE) to registry..."
	docker push $(LATEST_IMAGE)
	@echo "Successfully pushed latest tag"

# Kubernetes targets
setup-namespace: ## Create Kubernetes namespace
	kubectl apply -f k8s/namespace.yaml
	@echo "Namespace created/updated"

k8s-deploy: setup-namespace ## Deploy frontend to Kubernetes
	@echo "Deploying frontend to Kubernetes..."
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	@echo "Setting image to $(LATEST_IMAGE)..."
	kubectl set image deployment/godemode-frontend godemode-frontend=$(LATEST_IMAGE) -n godemode
	@echo "Waiting for rollout to complete..."
	kubectl rollout status deployment/godemode-frontend -n godemode
	@echo ""
	@echo "✅ Deployment complete!"
	@echo ""
	@echo "Check status with:"
	@echo "  kubectl get pods -n godemode"
	@echo "  kubectl get svc -n godemode"
	@echo ""
	@echo "Get service URL:"
	@kubectl get svc godemode-frontend -n godemode -o jsonpath='{.status.loadBalancer.ingress[0].ip}' && echo ""

deploy: docker-push k8s-deploy ## Full deployment: build, push, and deploy frontend to k8s
	@echo ""
	@echo "🚀 Frontend deployment complete!"
	@echo "Image: $(LATEST_IMAGE)"
	@echo ""
