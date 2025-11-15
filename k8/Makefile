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

.PHONY: build-backend
build-backend: ## Build backend Docker image
	@echo "Building backend image..."
	@echo "Tag: $(BACKEND_IMAGE):$(TAG)"
	docker build -f ../Dockerfile.backend -t $(BACKEND_IMAGE):$(TAG) ..
	docker tag $(BACKEND_IMAGE):$(TAG) $(BACKEND_IMAGE):latest
	@echo "✓ Backend image built successfully"

.PHONY: build-frontend
build-frontend: ## Build frontend Docker image
	@echo "Building frontend image..."
	@echo "Tag: $(FRONTEND_IMAGE):$(TAG)"
	docker build -f ../frontend/Dockerfile -t $(FRONTEND_IMAGE):$(TAG) ../frontend
	docker tag $(FRONTEND_IMAGE):$(TAG) $(FRONTEND_IMAGE):latest
	@echo "✓ Frontend image built successfully"

.PHONY: build
build: build-backend build-frontend ## Build all Docker images

.PHONY: push-backend
push-backend: ## Push backend image to registry
	@echo "Pushing backend image..."
	docker push $(BACKEND_IMAGE):$(TAG)
	docker push $(BACKEND_IMAGE):latest
	@echo "✓ Backend image pushed successfully"

.PHONY: push-frontend
push-frontend: ## Push frontend image to registry
	@echo "Pushing frontend image..."
	docker push $(FRONTEND_IMAGE):$(TAG)
	docker push $(FRONTEND_IMAGE):latest
	@echo "✓ Frontend image pushed successfully"

.PHONY: push
push: push-backend push-frontend ## Push all images to registry

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

.PHONY: apply-config
apply-config: setup-namespace ## Apply ConfigMap
	@echo "Applying ConfigMap..."
	kubectl apply -f configmap.yaml
	@echo "✓ ConfigMap applied successfully"

.PHONY: deploy-backend
deploy-backend: apply-config ## Deploy backend to Kubernetes
	@echo "Deploying backend..."
	kubectl apply -f backend-service.yaml
	kubectl apply -f backend-deployment.yaml
	kubectl apply -f backend-hpa.yaml
	@echo "Waiting for backend rollout..."
	kubectl rollout status deployment/godemode-backend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Backend deployed successfully"

.PHONY: deploy-frontend
deploy-frontend: ## Deploy frontend to Kubernetes
	@echo "Deploying frontend..."
	kubectl apply -f frontend-service.yaml
	kubectl apply -f frontend-deployment.yaml
	kubectl apply -f frontend-hpa.yaml
	@echo "Waiting for frontend rollout..."
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Frontend deployed successfully"

.PHONY: deploy-ingress
deploy-ingress: ## Deploy ingress
	@echo "Deploying ingress..."
	kubectl apply -f ingress.yaml
	@echo "✓ Ingress deployed successfully"

.PHONY: deploy
deploy: deploy-backend deploy-frontend deploy-ingress ## Deploy all components

.PHONY: full-deploy
full-deploy: check-context check-registry-secret setup-namespace apply-secrets apply-config deploy ## Complete deployment (namespace, secrets, config, all components)
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Full deployment complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@make status

.PHONY: update-backend
update-backend: build-backend push-backend ## Update backend (build, push, deploy)
	@echo "Updating backend deployment to use tag: $(TAG)"
	kubectl set image deployment/godemode-backend backend=$(BACKEND_IMAGE):$(TAG) -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-backend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Backend updated successfully"

.PHONY: update-frontend
update-frontend: build-frontend push-frontend ## Update frontend (build, push, deploy)
	@echo "Updating frontend deployment to use tag: $(TAG)"
	kubectl set image deployment/godemode-frontend frontend=$(FRONTEND_IMAGE):$(TAG) -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE) --timeout=300s
	@echo "✓ Frontend updated successfully"

.PHONY: update
update: update-backend update-frontend ## Update all components

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

.PHONY: logs-backend
logs-backend: ## Show backend logs
	kubectl logs -f deployment/godemode-backend -n $(NAMESPACE)

.PHONY: logs-frontend
logs-frontend: ## Show frontend logs
	kubectl logs -f deployment/godemode-frontend -n $(NAMESPACE)

.PHONY: describe-backend
describe-backend: ## Describe backend deployment
	kubectl describe deployment godemode-backend -n $(NAMESPACE)

.PHONY: describe-frontend
describe-frontend: ## Describe frontend deployment
	kubectl describe deployment godemode-frontend -n $(NAMESPACE)

.PHONY: restart-backend
restart-backend: ## Restart backend pods
	kubectl rollout restart deployment/godemode-backend -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-backend -n $(NAMESPACE)

.PHONY: restart-frontend
restart-frontend: ## Restart frontend pods
	kubectl rollout restart deployment/godemode-frontend -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-frontend -n $(NAMESPACE)

.PHONY: restart
restart: restart-backend restart-frontend ## Restart all pods

.PHONY: scale-backend
scale-backend: ## Scale backend (usage: make scale-backend REPLICAS=5)
	kubectl scale deployment godemode-backend --replicas=$(REPLICAS) -n $(NAMESPACE)
	@echo "✓ Backend scaled to $(REPLICAS) replicas"

.PHONY: scale-frontend
scale-frontend: ## Scale frontend (usage: make scale-frontend REPLICAS=3)
	kubectl scale deployment godemode-frontend --replicas=$(REPLICAS) -n $(NAMESPACE)
	@echo "✓ Frontend scaled to $(REPLICAS) replicas"

.PHONY: rollback-backend
rollback-backend: ## Rollback backend to previous version
	kubectl rollout undo deployment/godemode-backend -n $(NAMESPACE)
	kubectl rollout status deployment/godemode-backend -n $(NAMESPACE)
	@echo "✓ Backend rolled back"

.PHONY: rollback-frontend
rollback-frontend: ## Rollback frontend to previous version
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

.PHONY: shell-backend
shell-backend: ## Open shell in backend pod
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=godemode-backend -o jsonpath='{.items[0].metadata.name}'); \
	echo "Opening shell in pod: $$POD"; \
	kubectl exec -it $$POD -n $(NAMESPACE) -- /bin/sh

.PHONY: shell-frontend
shell-frontend: ## Open shell in frontend pod
	@POD=$$(kubectl get pod -n $(NAMESPACE) -l app=godemode-frontend -o jsonpath='{.items[0].metadata.name}'); \
	echo "Opening shell in pod: $$POD"; \
	kubectl exec -it $$POD -n $(NAMESPACE) -- /bin/sh

.PHONY: test-health
test-health: ## Test health endpoints
	@echo "Testing backend health..."
	@kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n $(NAMESPACE) -- \
		curl -s http://godemode-backend/health
	@echo ""
	@echo "Testing frontend health..."
	@kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n $(NAMESPACE) -- \
		curl -s http://godemode-frontend/health

# Quick deployment target - for CI/CD or fast deploys
.PHONY: quick-deploy
quick-deploy: build-push ## Quick build and deploy with latest tag
	kubectl set image deployment/godemode-backend backend=$(BACKEND_IMAGE):latest -n $(NAMESPACE)
	kubectl set image deployment/godemode-frontend frontend=$(FRONTEND_IMAGE):latest -n $(NAMESPACE)
	@echo "✓ Images updated, rollout in progress..."
