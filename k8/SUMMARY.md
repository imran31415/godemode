# Godemode Kubernetes Deployment - Setup Summary

## What Was Created

A complete Kubernetes deployment infrastructure for Godemode (backend + frontend) to deploy to the `godemode` namespace.

### Directory Structure

```
godemode/
├── Dockerfile.backend              # Backend Docker build
├── frontend/
│   └── Dockerfile                  # Frontend Docker build
└── k8/
    ├── namespace.yaml              # Namespace definition
    ├── configmap.yaml              # Environment variables
    ├── backend-deployment.yaml     # Backend deployment
    ├── backend-service.yaml        # Backend service
    ├── backend-hpa.yaml            # Backend auto-scaling
    ├── frontend-deployment.yaml    # Frontend deployment
    ├── frontend-service.yaml       # Frontend service
    ├── frontend-hpa.yaml           # Frontend auto-scaling
    ├── ingress.yaml                # Ingress with TLS
    ├── Makefile                    # Build & deploy automation
    ├── deploy.sh                   # Deployment script
    ├── README.md                   # Documentation
    ├── DEPLOYMENT_GUIDE.md         # Detailed guide
    └── secrets/
        ├── .gitignore              # Ignore secret files
        ├── README.md               # Secrets documentation
        └── backend-secrets.example.yaml  # Secret template
```

## Architecture

```
Internet
    │
    ▼
┌─────────────────────────────────────────────┐
│  Ingress (godemode.scalebase.io)            │
│  - TLS via cert-manager                     │
│  - Routes /api → backend                    │
│  - Routes / → frontend                      │
└─────────────────────────────────────────────┘
    │                    │
    ▼                    ▼
┌──────────────┐    ┌──────────────┐
│  Backend     │    │  Frontend    │
│  Service     │    │  Service     │
│  Port: 80    │    │  Port: 80    │
└──────────────┘    └──────────────┘
    │                    │
    ▼                    ▼
┌──────────────┐    ┌──────────────┐
│  Backend     │    │  Frontend    │
│  Deployment  │    │  Deployment  │
│  2-10 pods   │    │  2-8 pods    │
│  HPA enabled │    │  HPA enabled │
└──────────────┘    └──────────────┘
```

## Key Features

### 1. Multi-Stage Docker Builds
- **Backend**: Builds with Go 1.23, includes TinyGo for WASM compilation
- **Frontend**: Expo web build → Nginx static serving

### 2. Kubernetes Resources
- **Namespace**: Isolated `godemode` namespace
- **Deployments**: Separate backend and frontend deployments
- **Services**: ClusterIP services for internal communication
- **Ingress**: External access with TLS and path-based routing
- **HPA**: Auto-scaling based on CPU and memory

### 3. Configuration
- **ConfigMap**: Non-sensitive environment variables
- **Secrets**: Sensitive data (ANTHROPIC_API_KEY, etc.)
- **Registry**: Uses DigitalOcean Container Registry

### 4. Automation
- **Makefile**: 30+ commands for build, deploy, monitor, troubleshoot
- **deploy.sh**: Interactive deployment script

## Quick Start

```bash
# 1. Navigate to k8 directory
cd k8

# 2. Setup secrets
make setup-secrets
# Edit secrets/backend-secrets.yaml with your ANTHROPIC_API_KEY

# 3. Deploy everything
make full-deploy
```

## Common Commands

```bash
# Build & Deploy
make build              # Build Docker images
make push               # Push to registry
make deploy             # Deploy to Kubernetes
make full-deploy        # Complete deployment

# Update
make update             # Update all after code changes
make update-backend     # Update backend only
make update-frontend    # Update frontend only

# Monitor
make status             # Show deployment status
make logs-backend       # View backend logs
make logs-frontend      # View frontend logs

# Manage
make restart            # Restart all pods
make rollback-backend   # Rollback backend
make scale-backend REPLICAS=5  # Scale manually

# Troubleshoot
make shell-backend      # Open shell in pod
make test-health        # Test health endpoints
make describe-backend   # Describe deployment

# Cleanup
make clean              # Delete namespace
```

## Configuration Summary

### Backend
- **Image**: `registry.digitalocean.com/resourceloop/godemode/backend:latest`
- **Port**: 8080
- **Health**: `/health`
- **Replicas**: 2-10 (HPA)
- **Resources**: 256Mi-512Mi RAM, 100m-500m CPU

### Frontend
- **Image**: `registry.digitalocean.com/resourceloop/godemode/frontend:latest`
- **Port**: 80
- **Health**: `/`
- **Replicas**: 2-8 (HPA)
- **Resources**: 128Mi-256Mi RAM, 50m-200m CPU

### Ingress
- **Domain**: godemode.scalebase.io
- **TLS**: Let's Encrypt (auto)
- **Routes**:
  - `/api` → backend:80
  - `/health` → backend:80
  - `/` → frontend:80

## Environment Variables

### ConfigMap (configmap.yaml)
- `PORT`: 8080
- `ENVIRONMENT`: production
- `API_URL`: https://godemode.scalebase.io/api
- `TINYGO_ROOT`: /usr/local/tinygo
- `DEFAULT_TIMEOUT`: 30s
- `DEFAULT_MEMORY_MB`: 64
- `MAX_MEMORY_MB`: 256
- `CACHE_ENABLED`: true
- `LOG_LEVEL`: info

### Secrets (secrets/backend-secrets.yaml)
- `ANTHROPIC_API_KEY`: Your API key (required)

## Prerequisites

✅ Docker installed and running
✅ kubectl configured with cluster access
✅ Registry secrets in `fulcrum` or `agentlog` namespace
✅ cert-manager installed in cluster
✅ nginx-ingress-controller installed

## Security Features

- ✅ Secrets not committed to git (gitignored)
- ✅ Registry authentication via imagePullSecrets
- ✅ TLS/SSL via cert-manager
- ✅ Health checks for liveness and readiness
- ✅ Resource limits to prevent resource exhaustion
- ✅ Non-root containers (Alpine base)

## Deployment Flow

1. **Build**: Multi-stage Docker builds for both services
2. **Push**: Push to DigitalOcean registry
3. **Setup**: Create namespace, copy registry secrets
4. **Config**: Apply ConfigMap and Secrets
5. **Deploy**: Deploy services, deployments, HPAs
6. **Ingress**: Configure external access
7. **Monitor**: Check status and logs

## Registry Secret Management

The Makefile automatically:
- Checks for existing registry secrets in `fulcrum` or `agentlog` namespaces
- Copies them to the `godemode` namespace
- Uses the same credentials as other projects

## Next Steps

1. **Configure Secrets**:
   ```bash
   cd k8
   make setup-secrets
   vim secrets/backend-secrets.yaml
   # Add your ANTHROPIC_API_KEY
   ```

2. **Deploy**:
   ```bash
   make full-deploy
   ```

3. **Verify**:
   ```bash
   make status
   make logs-backend
   ```

4. **Access**:
   - Open: https://godemode.scalebase.io

## Troubleshooting

### Pods Not Starting
```bash
kubectl get pods -n godemode
kubectl describe pod <pod-name> -n godemode
kubectl logs <pod-name> -n godemode
```

### Image Pull Errors
```bash
make setup-namespace  # Recreate registry secret
```

### Ingress Issues
```bash
kubectl describe ingress godemode-ingress -n godemode
kubectl get certificate -n godemode
```

### Application Errors
```bash
make logs-backend
make logs-frontend
make test-health
```

## Documentation

- **README.md**: Overview and architecture
- **DEPLOYMENT_GUIDE.md**: Detailed deployment instructions
- **secrets/README.md**: Secrets management
- **Makefile**: Self-documenting (run `make help`)

## References

Based on deployment patterns from:
- `../fulcrum/k8`: Single-service web app deployment
- `../agentlog/k8s`: Multi-service (backend + frontend) deployment

## Support

For deployment issues:
1. Run `make status` to check current state
2. Run `make logs-backend` or `make logs-frontend` for logs
3. Check events: `kubectl get events -n godemode`
4. Use `make help` to see all available commands

## Success Criteria

✅ Namespace created
✅ Registry secrets configured
✅ Backend secrets applied
✅ ConfigMap applied
✅ Backend deployed and running (2+ pods)
✅ Frontend deployed and running (2+ pods)
✅ Services accessible internally
✅ Ingress configured with TLS
✅ Application accessible at https://godemode.scalebase.io
✅ Health checks passing
✅ HPA monitoring and ready to scale

---

**Created**: November 15, 2025
**Cluster**: Same as fulcrum/agentlog
**Namespace**: godemode
**Domain**: godemode.scalebase.io
**Registry**: registry.digitalocean.com/resourceloop/godemode
