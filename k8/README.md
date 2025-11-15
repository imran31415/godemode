# Godemode Kubernetes Deployment

This directory contains Kubernetes manifests for deploying Godemode to a Kubernetes cluster. The setup includes both backend (Go) and frontend (React Native Web) components deployed to the `godemode` namespace.

## Architecture

```
┌─────────────────────────────────────────────┐
│          Ingress (godemode.scalebase.io)    │
│  ┌──────────────────────────────────────┐   │
│  │ TLS: letsencrypt-production          │   │
│  └──────────────────────────────────────┘   │
└────────┬────────────────────┬───────────────┘
         │                    │
         │ /api, /health      │ /
         ▼                    ▼
┌─────────────────┐   ┌─────────────────┐
│  Backend        │   │  Frontend       │
│  Service        │   │  Service        │
│  (ClusterIP)    │   │  (ClusterIP)    │
│  Port: 80       │   │  Port: 80       │
└────────┬────────┘   └────────┬────────┘
         │                     │
         ▼                     ▼
┌─────────────────┐   ┌─────────────────┐
│  Backend        │   │  Frontend       │
│  Deployment     │   │  Deployment     │
│  - Replicas: 2+ │   │  - Replicas: 2+ │
│  - HPA enabled  │   │  - HPA enabled  │
│  - Port: 8080   │   │  - Port: 80     │
└─────────────────┘   └─────────────────┘
```

## Components

### Core Resources

- **namespace.yaml**: Creates the `godemode` namespace
- **configmap.yaml**: Environment variables for both backend and frontend
- **backend-deployment.yaml**: Backend Go application deployment
- **backend-service.yaml**: Backend ClusterIP service
- **frontend-deployment.yaml**: Frontend React Native Web deployment
- **frontend-service.yaml**: Frontend ClusterIP service
- **ingress.yaml**: Nginx ingress for external access with TLS

### Auto-scaling

- **backend-hpa.yaml**: Horizontal Pod Autoscaler for backend (2-10 replicas)
- **frontend-hpa.yaml**: Horizontal Pod Autoscaler for frontend (2-8 replicas)

### Scripts

- **deploy.sh**: Automated deployment script

## Prerequisites

1. **Kubernetes cluster** with kubectl access
2. **Nginx Ingress Controller** installed
3. **cert-manager** for automatic TLS certificates
4. **DigitalOcean Container Registry** access (or modify for your registry)
5. **Docker images** built and pushed to registry:
   - `registry.digitalocean.com/resourceloop/godemode/backend:latest`
   - `registry.digitalocean.com/resourceloop/godemode/frontend:latest`

## Configuration

### Image Registry

The manifests are configured for DigitalOcean Container Registry. Update the image references in:
- `backend-deployment.yaml` (line 24)
- `frontend-deployment.yaml` (line 24)

### Domain

Default domain is `godemode.scalebase.io`. Update in:
- `ingress.yaml` (lines 13, 21, 28)
- `configmap.yaml` (line 9)

### Secrets

For sensitive data (API keys, database credentials, etc.), create a `secrets/` directory:

```bash
mkdir -p secrets
```

Example secret file (`secrets/backend-secrets.yaml`):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: godemode-backend-secrets
  namespace: godemode
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "your-api-key-here"
  DATABASE_URL: "your-database-url"
  # Add other secrets
```

**Important**: Add `secrets/` to `.gitignore` to avoid committing sensitive data.

### Registry Credentials

Create a secret for pulling images from the registry:

```bash
kubectl create secret docker-registry digitalocean-registry \
  --docker-server=registry.digitalocean.com \
  --docker-username=your-username \
  --docker-password=your-password \
  --namespace=godemode
```

## Deployment

### Quick Deploy

```bash
cd k8
./deploy.sh
```

### Manual Deploy

```bash
# Apply resources in order
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets/  # if you have secrets
kubectl apply -f backend-service.yaml
kubectl apply -f frontend-service.yaml
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f backend-hpa.yaml
kubectl apply -f frontend-hpa.yaml
kubectl apply -f ingress.yaml
```

### Verify Deployment

```bash
# Check all resources
kubectl get all -n godemode

# Check pods
kubectl get pods -n godemode

# Check services
kubectl get svc -n godemode

# Check ingress
kubectl get ingress -n godemode

# Check HPAs
kubectl get hpa -n godemode
```

## Monitoring

### View Logs

```bash
# Backend logs
kubectl logs -f deployment/godemode-backend -n godemode

# Frontend logs
kubectl logs -f deployment/godemode-frontend -n godemode

# Specific pod logs
kubectl logs -f <pod-name> -n godemode
```

### Pod Status

```bash
# Get pod details
kubectl describe pod <pod-name> -n godemode

# Get deployment status
kubectl rollout status deployment/godemode-backend -n godemode
kubectl rollout status deployment/godemode-frontend -n godemode
```

### Scaling

```bash
# Manual scaling
kubectl scale deployment godemode-backend --replicas=5 -n godemode

# Check HPA status
kubectl get hpa -n godemode

# Describe HPA for details
kubectl describe hpa godemode-backend-hpa -n godemode
```

## Updating

### Update Backend

```bash
# Build and push new image
docker build -t registry.digitalocean.com/resourceloop/godemode/backend:v1.1.0 .
docker push registry.digitalocean.com/resourceloop/godemode/backend:v1.1.0

# Update deployment
kubectl set image deployment/godemode-backend \
  backend=registry.digitalocean.com/resourceloop/godemode/backend:v1.1.0 \
  -n godemode

# Or edit the deployment manifest and re-apply
kubectl apply -f backend-deployment.yaml
```

### Update Frontend

```bash
# Build and push new image
docker build -t registry.digitalocean.com/resourceloop/godemode/frontend:v1.1.0 ./frontend
docker push registry.digitalocean.com/resourceloop/godemode/frontend:v1.1.0

# Update deployment
kubectl set image deployment/godemode-frontend \
  frontend=registry.digitalocean.com/resourceloop/godemode/frontend:v1.1.0 \
  -n godemode
```

### Rollback

```bash
# View rollout history
kubectl rollout history deployment/godemode-backend -n godemode

# Rollback to previous version
kubectl rollout undo deployment/godemode-backend -n godemode

# Rollback to specific revision
kubectl rollout undo deployment/godemode-backend --to-revision=2 -n godemode
```

## Health Checks

The deployments include health checks:

### Backend
- **Readiness**: `GET /health` on port 8080
  - Initial delay: 10s
  - Period: 10s
- **Liveness**: `GET /health` on port 8080
  - Initial delay: 20s
  - Period: 20s

### Frontend
- **Readiness**: `GET /` on port 80
  - Initial delay: 10s
  - Period: 5s
- **Liveness**: `GET /` on port 80
  - Initial delay: 30s
  - Period: 10s

## Resource Limits

### Backend
- **Requests**: 256Mi memory, 100m CPU
- **Limits**: 512Mi memory, 500m CPU

### Frontend
- **Requests**: 128Mi memory, 50m CPU
- **Limits**: 256Mi memory, 200m CPU

Adjust these in the deployment manifests based on your needs.

## Troubleshooting

### Pods not starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n godemode

# Check logs
kubectl logs <pod-name> -n godemode
```

### Image pull errors

```bash
# Verify registry secret
kubectl get secret digitalocean-registry -n godemode

# Recreate if needed
kubectl delete secret digitalocean-registry -n godemode
kubectl create secret docker-registry digitalocean-registry \
  --docker-server=registry.digitalocean.com \
  --docker-username=your-username \
  --docker-password=your-password \
  --namespace=godemode
```

### Ingress not working

```bash
# Check ingress status
kubectl describe ingress godemode-ingress -n godemode

# Check nginx controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Verify cert-manager
kubectl get certificate -n godemode
kubectl describe certificate godemode-tls -n godemode
```

### HPA not scaling

```bash
# Check metrics server
kubectl top nodes
kubectl top pods -n godemode

# If metrics unavailable, install metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## Clean Up

### Delete all resources

```bash
kubectl delete namespace godemode
```

### Delete specific resources

```bash
kubectl delete -f ingress.yaml
kubectl delete -f backend-deployment.yaml
kubectl delete -f frontend-deployment.yaml
# etc.
```

## Security Considerations

1. **Secrets**: Never commit secrets to version control
2. **Registry**: Use private registry with authentication
3. **RBAC**: Implement proper Role-Based Access Control
4. **Network Policies**: Consider adding network policies to restrict traffic
5. **Pod Security**: Enable pod security standards
6. **TLS**: Always use HTTPS in production (configured via ingress)

## References

- Based on deployment patterns from `../fulcrum/k8` and `../agentlog/k8s`
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager](https://cert-manager.io/)
- [DigitalOcean Kubernetes](https://docs.digitalocean.com/products/kubernetes/)
