# Godemode Deployment Guide

Complete guide for deploying Godemode to Kubernetes using the provided Makefile and manifests.

## Quick Start

For a complete first-time deployment:

```bash
cd k8

# 1. Setup secrets
make setup-secrets
# Edit secrets/backend-secrets.yaml with your ANTHROPIC_API_KEY

# 2. Deploy everything
make full-deploy
```

## Prerequisites

- Docker installed and running
- kubectl configured with access to your cluster
- Access to DigitalOcean Container Registry (registry.digitalocean.com)
- Registry secrets already exist in `fulcrum` or `agentlog` namespaces

## Step-by-Step Deployment

### 1. Verify Prerequisites

```bash
# Check kubectl context
kubectl config current-context

# Verify registry secrets exist
kubectl get secret -n fulcrum digitalocean-registry
# OR
kubectl get secret -n agentlog umi-backend
```

### 2. Setup Secrets

```bash
cd k8

# Create secrets file from template
make setup-secrets

# Edit the secrets file with your actual values
vim secrets/backend-secrets.yaml
# At minimum, set ANTHROPIC_API_KEY
```

### 3. Build Docker Images

```bash
# Build both backend and frontend
make build

# Or build individually
make build-backend
make build-frontend
```

This will create images:
- `registry.digitalocean.com/resourceloop/godemode/backend:latest`
- `registry.digitalocean.com/resourceloop/godemode/frontend:latest`

### 4. Push to Registry

```bash
# Login to DigitalOcean registry first
doctl registry login

# Push both images
make push

# Or push individually
make push-backend
make push-frontend
```

### 5. Deploy to Kubernetes

```bash
# Full deployment (recommended for first time)
make full-deploy

# This will:
# - Check kubectl context (asks for confirmation)
# - Verify registry secrets exist
# - Create godemode namespace
# - Copy registry secrets to godemode namespace
# - Apply backend secrets
# - Apply ConfigMap
# - Deploy backend (deployment, service, HPA)
# - Deploy frontend (deployment, service, HPA)
# - Deploy ingress
# - Show deployment status
```

### 6. Verify Deployment

```bash
# Check status
make status

# Watch pods come up
kubectl get pods -n godemode -w

# Check logs
make logs-backend
# OR
make logs-frontend
```

## Makefile Commands

### Building & Pushing

```bash
make build              # Build both images
make build-backend      # Build backend only
make build-frontend     # Build frontend only
make push               # Push both images
make push-backend       # Push backend only
make push-frontend      # Push frontend only
make build-push         # Build and push everything
```

### Deployment

```bash
make full-deploy        # Complete deployment (first time)
make deploy             # Deploy all components (assumes namespace exists)
make deploy-backend     # Deploy backend only
make deploy-frontend    # Deploy frontend only
make deploy-ingress     # Deploy ingress only
```

### Updates

```bash
# Update after code changes
make update             # Update both backend and frontend
make update-backend     # Build, push, and update backend
make update-frontend    # Build, push, and update frontend

# Quick deploy (build & push with latest tag, update deployments)
make quick-deploy
```

### Configuration

```bash
make setup-namespace    # Create namespace and copy registry secret
make apply-secrets      # Apply secrets to cluster
make apply-config       # Apply ConfigMap
```

### Monitoring

```bash
make status             # Show deployment status
make logs-backend       # Tail backend logs
make logs-frontend      # Tail frontend logs
make describe-backend   # Describe backend deployment
make describe-frontend  # Describe frontend deployment
```

### Scaling

```bash
# Manual scaling
make scale-backend REPLICAS=5
make scale-frontend REPLICAS=3

# HPA will auto-scale based on CPU/memory
```

### Restart

```bash
make restart            # Restart both deployments
make restart-backend    # Restart backend pods
make restart-frontend   # Restart frontend pods
```

### Rollback

```bash
make rollback-backend   # Rollback backend to previous version
make rollback-frontend  # Rollback frontend to previous version
```

### Troubleshooting

```bash
make shell-backend      # Open shell in backend pod
make shell-frontend     # Open shell in frontend pod
make test-health        # Test health endpoints
```

### Cleanup

```bash
make clean              # Delete entire namespace (asks for confirmation)
make clean-images       # Clean local Docker images
```

### Help

```bash
make help               # Show all available commands
```

## Common Workflows

### First Time Deployment

```bash
cd k8
make setup-secrets
# Edit secrets/backend-secrets.yaml
make full-deploy
```

### Update After Code Changes

```bash
cd k8
make update
# This builds, pushes, and updates deployments
```

### Check Application Status

```bash
cd k8
make status
```

### View Logs

```bash
cd k8
make logs-backend
# Press Ctrl+C to exit, then:
make logs-frontend
```

### Rollback Bad Deployment

```bash
cd k8
make rollback-backend
# OR
make rollback-frontend
```

### Scale for High Traffic

```bash
cd k8
# Manual scaling
make scale-backend REPLICAS=10

# Or let HPA handle it automatically
# HPA will scale 2-10 replicas based on CPU/memory
```

## Environment Variables

### Backend (via ConfigMap)

Set in `configmap.yaml`:
- `PORT`: Backend port (default: 8080)
- `ENVIRONMENT`: Environment name (production)
- `API_URL`: API base URL
- `TINYGO_ROOT`: TinyGo installation path
- `DEFAULT_TIMEOUT`: Default execution timeout
- `DEFAULT_MEMORY_MB`: Default memory limit
- `MAX_MEMORY_MB`: Maximum memory limit
- `CACHE_ENABLED`: Enable module caching
- `LOG_LEVEL`: Logging level

### Backend (via Secrets)

Set in `secrets/backend-secrets.yaml`:
- `ANTHROPIC_API_KEY`: Required - Your Anthropic API key

## Image Tagging

By default, images are tagged with:
- Git commit hash + timestamp (e.g., `abc1234-20231115-143000`)
- `latest` tag

Override the tag:
```bash
make build TAG=v1.2.0
make push TAG=v1.2.0
```

## Custom Registry

To use a different registry, set the `REGISTRY` variable:

```bash
make build REGISTRY=your-registry.com/project
make push REGISTRY=your-registry.com/project
```

## Accessing the Application

Once deployed, the application is accessible at:
- **Production**: https://godemode.scalebase.io

Routes:
- `/` - Frontend (React Native Web)
- `/api` - Backend API
- `/health` - Backend health check

## TLS/SSL

TLS is automatically configured via:
- cert-manager with Let's Encrypt
- Annotation: `cert-manager.io/cluster-issuer: "letsencrypt-production"`
- Secret: `godemode-tls` (auto-generated)

## Auto-scaling

Both backend and frontend have HPA configured:

**Backend**: 2-10 replicas
- Scale on CPU: 70%
- Scale on Memory: 80%

**Frontend**: 2-8 replicas
- Scale on CPU: 70%
- Scale on Memory: 80%

View HPA status:
```bash
kubectl get hpa -n godemode
kubectl describe hpa godemode-backend-hpa -n godemode
```

## Health Checks

### Backend
- **Readiness**: `GET /health` every 10s (delay 10s)
- **Liveness**: `GET /health` every 20s (delay 20s)

### Frontend
- **Readiness**: `GET /` every 5s (delay 10s)
- **Liveness**: `GET /` every 10s (delay 30s)

## Resource Limits

### Backend
- Requests: 256Mi memory, 100m CPU
- Limits: 512Mi memory, 500m CPU

### Frontend
- Requests: 128Mi memory, 50m CPU
- Limits: 256Mi memory, 200m CPU

Adjust in deployment manifests as needed.

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n godemode

# Describe pod
kubectl describe pod <pod-name> -n godemode

# Check events
kubectl get events -n godemode --sort-by='.lastTimestamp'
```

### Image Pull Errors

```bash
# Verify registry secret
kubectl get secret digitalocean-registry -n godemode

# Recreate if needed
make setup-namespace
```

### Application Errors

```bash
# Check logs
make logs-backend
make logs-frontend

# Check environment variables
kubectl exec -n godemode <pod-name> -- env
```

### Ingress Not Working

```bash
# Check ingress
kubectl describe ingress godemode-ingress -n godemode

# Check certificate
kubectl get certificate -n godemode
kubectl describe certificate godemode-tls -n godemode

# Check nginx logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
```

### Performance Issues

```bash
# Check resource usage
kubectl top pods -n godemode
kubectl top nodes

# Check HPA
kubectl get hpa -n godemode
kubectl describe hpa godemode-backend-hpa -n godemode
```

## CI/CD Integration

### Example GitHub Actions

```yaml
name: Deploy to Kubernetes

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build and Deploy
        run: |
          cd k8
          make build-push
          make update
        env:
          KUBECONFIG: ${{ secrets.KUBECONFIG }}
```

## Security Best Practices

1. **Never commit secrets** - Use the provided `.gitignore` in `secrets/`
2. **Use RBAC** - Limit kubectl access to namespace
3. **Scan images** - Use security scanning tools
4. **Update regularly** - Keep dependencies updated
5. **Monitor logs** - Watch for security events
6. **Use network policies** - Restrict pod-to-pod traffic
7. **Rotate secrets** - Regularly update API keys

## Support

For issues or questions:
- Check logs: `make logs-backend` / `make logs-frontend`
- Check status: `make status`
- Verify health: `make test-health`
- Review events: `kubectl get events -n godemode`
