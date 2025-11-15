# Godemode Kubernetes Build and Deploy Makefile

# Configuration
REGISTRY := registry.digitalocean.com/resourceloop
PROJECT := godemode
NAMESPACE := godemode

# Image names
BACKEND_IMAGE := $(REGISTRY)/$(PROJECT)/backend
FRONTEND_IMAGE := $(REGISTRY)/$(PROJECT)/frontend

# Default tag (use git commit hash or 'latest')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "latest")
TIMESTAMP := $(shell date +%Y%m%d-%H%M%S)
TAG ?= $(GIT_COMMIT)-$(TIMESTAMP)

# Kubernetes context check
KUBE_CONTEXT := $(shell kubectl config current-context 2>/dev/null)

.PHONY: help
help: ## Show this help message
	@echo "Godemode Kubernetes Deployment Makefile"
	@echo "========================================"
	@echo ""
	@echo "Current Configuration:"
	@echo "  Registry:        $(REGISTRY)"
	@echo "  Project:         $(PROJECT)"
	@echo "  Namespace:       $(NAMESPACE)"
	@echo "  Tag:             $(TAG)"
	@echo "  Kube Context:    $(KUBE_CONTEXT)"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: check-context
check-context: ## Check kubectl context
	@echo "Current kubectl context: $(KUBE_CONTEXT)"
	@read -p "Continue with this context? [y/N] " -n 1 -r; \
	echo; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Aborted. Change context with: kubectl config use-context <context-name>"; \
		exit 1; \
	fi

.PHONY: check-registry-secret
check-registry-secret: ## Check if registry secret exists
	@echo "Checking for registry secret..."
	@if kubectl get secret -n fulcrum digitalocean-registry >/dev/null 2>&1; then \
		echo "✓ Found digitalocean-registry secret in fulcrum namespace"; \
	elif kubectl get secret -n agentlog umi-backend >/dev/null 2>&1; then \
		echo "✓ Found umi-backend secret in agentlog namespace"; \
	else \
		echo "✗ No registry secret found in fulcrum or agentlog namespaces"; \
		echo "Please ensure registry secrets exist first"; \
		exit 1; \
	fi

.PHONY: setup-namespace
setup-namespace: ## Create namespace and copy registry secret
	@echo "Setting up namespace..."
	kubectl apply -f namespace.yaml
	@echo "Copying registry secret from fulcrum namespace..."
	@if kubectl get secret -n fulcrum digitalocean-registry >/dev/null 2>&1; then \
		kubectl get secret digitalocean-registry -n fulcrum -o yaml | \
		sed 's/namespace: fulcrum/namespace: $(NAMESPACE)/' | \
		kubectl apply -f -; \
		echo "✓ Registry secret copied successfully"; \
	elif kubectl get secret -n agentlog umi-backend >/dev/null 2>&1; then \
		kubectl get secret umi-backend -n agentlog -o yaml | \
		sed 's/namespace: agentlog/namespace: $(NAMESPACE)/' | \
		sed 's/name: umi-backend/name: digitalocean-registry/' | \
		kubectl apply -f -; \
		echo "✓ Registry secret copied successfully"; \
	else \
		echo "✗ No registry secret found to copy"; \
		exit 1; \
	fi

.PHONY: build-frontend
build-frontend: ## Build frontend Docker image for amd64
	@echo "Building frontend image for linux/amd64..."
	@echo "Tag: $(FRONTEND_IMAGE):$(TAG)"
	docker buildx build --platform linux/amd64 -f ../frontend/Dockerfile -t $(FRONTEND_IMAGE):$(TAG) -t $(FRONTEND_IMAGE):latest ../frontend
	@echo "✓ Frontend image built successfully"

.PHONY: build
build: build-frontend ## Build frontend Docker image

.PHONY: push-frontend
push-frontend: ## Push frontend image to registry
	@echo "Pushing frontend image..."
	docker push $(FRONTEND_IMAGE):$(TAG)
	docker push $(FRONTEND_IMAGE):latest
	@echo "✓ Frontend image pushed successfully"

.PHONY: push
push: push-frontend ## Push frontend image to registry

.PHONY: build-push
build-push: build push ## Build and push all images

.PHONY: setup-secrets
setup-secrets: ## Setup secrets (interactive)
	@echo "Setting up secrets..."
	@if [ ! -f secrets/backend-secrets.yaml ]; then \
		echo "Creating backend-secrets.yaml from example..."; \
		cp secrets/backend-secrets.example.yaml secrets/backend-secrets.yaml; \
		echo ""; \
		echo "⚠️  Please edit secrets/backend-secrets.yaml with your actual secrets"; \
		echo "   Then run: make apply-secrets"; \
		exit 1; \
	else \
		echo "✓ secrets/backend-secrets.yaml exists"; \
	fi

.PHONY: apply-secrets
apply-secrets: setup-namespace ## Apply secrets to cluster
	@echo "Applying secrets..."
	@if [ -f secrets/backend-secrets.yaml ]; then \
		kubectl apply -f secrets/backend-secrets.yaml; \
		echo "✓ Secrets applied successfully"; \
	else \
		echo "✗ secrets/backend-secrets.yaml not found"; \
		echo "   Run: make setup-secrets"; \
		exit 1; \
	fi

.PHONY: deploy-k8s
deploy-k8s: setup-namespace ## Deploy frontend to Kubernetes
	@echo "Deploying frontend to Kubernetes..."
	kubectl apply -f frontend-service.yaml
	kubectl apply -f frontend-deployment.yaml
	kubectl apply -f frontend-hpa.yaml
	kubectl apply -f ingress.yaml
	@echo "Waiting for frontend rollout..."
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Frontend deployed successfully"

.PHONY: deploy
deploy: check-context build push deploy-k8s ## Build, push, and deploy frontend (complete workflow)
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Deployment complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@make status

.PHONY: update-deployment
update-deployment: ## Update existing deployment with latest image
	@echo "Updating frontend deployment to use tag: $(TAG)"
	kubectl set image deployment/godemode-frontend frontend=$(FRONTEND_IMAGE):latest -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Frontend updated successfully"

.PHONY: update
update: build push update-deployment ## Update frontend (build, push, deploy)

.PHONY: status
status: ## Show deployment status
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📊 Deployment Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Pods:"
	@kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Services:"
	@kubectl get services -n $(NAMESPACE)
	@echo ""
	@echo "Ingress:"
	@kubectl get ingress -n $(NAMESPACE)
	@echo ""
	@echo "HPAs:"
	@kubectl get hpa -n $(NAMESPACE)
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🌐 Application URL: https://godemode.scalebase.io"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

.PHONY: logs
logs: ## Show frontend logs
	kubectl logs -f deployment/godemode-frontend -n $(NAMESPACE)

.PHONY: describe
describe: ## Describe frontend deployment
	kubectl describe deployment godemode-frontend -n $(NAMESPACE)

.PHONY: restart
restart: ## Restart frontend pods
	kubectl rollout restart deployment/godemode-frontend -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE)

.PHONY: scale
scale: ## Scale frontend (usage: make scale REPLICAS=3)
	kubectl scale deployment godemode-frontend --replicas=$(REPLICAS) -n $(NAMESPACE)
	@echo "✓ Frontend scaled to $(REPLICAS) replicas"

.PHONY: rollback
rollback: ## Rollback frontend to previous version
	kubectl rollout undo deployment/godemode-frontend -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE)
	@echo "✓ Frontend rolled back"

.PHONY: clean
clean: ## Delete all resources in namespace
	@echo "⚠️  This will delete all resources in namespace $(NAMESPACE)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		kubectl delete namespace $(NAMESPACE); \
		echo "✓ Namespace $(NAMESPACE) deleted"; \
	else \
		echo "Cancelled"; \
	fi

.PHONY: clean-images
clean-images: ## Clean local Docker images
	@echo "Cleaning local Docker images..."
	docker rmi $(BACKEND_IMAGE):$(TAG) $(BACKEND_IMAGE):latest || true
	docker rmi $(FRONTEND_IMAGE):$(TAG) $(FRONTEND_IMAGE):latest || true
	@echo "✓ Local images cleaned"

.PHONY: shell
shell: ## Open shell in frontend pod
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=godemode-frontend -o jsonpath='{.items[0].metadata.name}'); \
	echo "Opening shell in pod: $$POD"; \
	kubectl exec -it $$POD -n $(NAMESPACE) -- /bin/sh

.PHONY: test-health
test-health: ## Test health endpoints
	@echo "Testing frontend health..."
	@kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n $(NAMESPACE) -- \
		curl -s http://godemode-frontend/health
