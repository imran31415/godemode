# Godemode K8s - Quick Reference Card

## First Time Setup (5 min)

```bash
cd k8
make setup-secrets              # Creates secrets/backend-secrets.yaml
vim secrets/backend-secrets.yaml  # Add ANTHROPIC_API_KEY
make full-deploy                # Deploy everything
```

## Daily Commands

### Deploy & Update
```bash
make full-deploy      # Complete deployment
make update           # Update after code changes
make update-backend   # Update backend only
make update-frontend  # Update frontend only
make quick-deploy     # Fast deploy with latest tag
```

### Monitor
```bash
make status           # Show everything
make logs-backend     # Tail backend logs
make logs-frontend    # Tail frontend logs
```

### Manage
```bash
make restart          # Restart all
make restart-backend  # Restart backend
make scale-backend REPLICAS=5  # Scale manually
make rollback-backend # Rollback to previous
```

### Debug
```bash
make shell-backend    # Open shell in pod
make test-health      # Test health endpoints
make describe-backend # Detailed pod info
kubectl get events -n godemode  # Recent events
```

## Image Management

```bash
make build            # Build both images
make push             # Push both images
make build-push       # Build and push both

# With custom tag
make build TAG=v1.2.0
make push TAG=v1.2.0
```

## Kubectl Shortcuts

```bash
# Quick aliases (add to ~/.bashrc)
alias kgp='kubectl get pods -n godemode'
alias kgs='kubectl get svc -n godemode'
alias kgi='kubectl get ingress -n godemode'
alias kgh='kubectl get hpa -n godemode'
alias kl='kubectl logs -f -n godemode'
alias kd='kubectl describe -n godemode'
```

## Common Issues

| Issue | Solution |
|-------|----------|
| Pods not starting | `kubectl describe pod <name> -n godemode` |
| Image pull failed | `make setup-namespace` |
| App not accessible | `kubectl describe ingress -n godemode` |
| Need to check logs | `make logs-backend` |
| Pod crash loop | `kubectl logs <pod> -n godemode --previous` |

## URLs

- **Production**: https://godemode.scalebase.io
- **Backend API**: https://godemode.scalebase.io/api
- **Health Check**: https://godemode.scalebase.io/health

## Registry

```bash
# Login
doctl registry login

# Images
registry.digitalocean.com/resourceloop/godemode/backend:latest
registry.digitalocean.com/resourceloop/godemode/frontend:latest
```

## Namespace: godemode

```bash
# All resources
kubectl get all -n godemode

# Specific resources
kubectl get pods -n godemode
kubectl get svc -n godemode
kubectl get ingress -n godemode
kubectl get hpa -n godemode
kubectl get secrets -n godemode
kubectl get configmap -n godemode
```

## Resource Limits

| Component | Requests | Limits |
|-----------|----------|--------|
| Backend | 256Mi / 100m | 512Mi / 500m |
| Frontend | 128Mi / 50m | 256Mi / 200m |

## Auto-Scaling

| Component | Min | Max | Triggers |
|-----------|-----|-----|----------|
| Backend | 2 | 10 | CPU 70%, Mem 80% |
| Frontend | 2 | 8 | CPU 70%, Mem 80% |

## Health Checks

- **Backend**: `GET /health` on 8080
- **Frontend**: `GET /` on 80

## Emergency Commands

```bash
# Delete everything (asks for confirmation)
make clean

# Force restart
kubectl delete pod -l app=godemode-backend -n godemode
kubectl delete pod -l app=godemode-frontend -n godemode

# Check cluster resources
kubectl top nodes
kubectl top pods -n godemode

# View all events
kubectl get events -n godemode --sort-by='.lastTimestamp'
```

## Configuration Files

| File | Purpose |
|------|---------|
| `namespace.yaml` | Creates godemode namespace |
| `configmap.yaml` | Non-sensitive env vars |
| `secrets/backend-secrets.yaml` | API keys, secrets |
| `backend-deployment.yaml` | Backend pods config |
| `backend-service.yaml` | Backend service |
| `backend-hpa.yaml` | Backend auto-scaling |
| `frontend-deployment.yaml` | Frontend pods config |
| `frontend-service.yaml` | Frontend service |
| `frontend-hpa.yaml` | Frontend auto-scaling |
| `ingress.yaml` | External access & TLS |

## Makefile Targets (Full List)

Run `make help` for complete list. Key targets:

- `help` - Show help
- `build` - Build images
- `push` - Push images
- `deploy` - Deploy to k8s
- `full-deploy` - Complete deployment
- `update` - Update all
- `status` - Show status
- `logs-backend` - Backend logs
- `logs-frontend` - Frontend logs
- `restart` - Restart all
- `rollback-backend` - Rollback
- `scale-backend` - Scale manually
- `shell-backend` - Open shell
- `test-health` - Test health
- `clean` - Delete namespace

## Workflow Examples

### Update Backend Code
```bash
# Edit code...
cd k8
make update-backend
make logs-backend  # Verify
```

### Emergency Rollback
```bash
cd k8
make rollback-backend
make status
```

### Scale for Traffic
```bash
cd k8
make scale-backend REPLICAS=10
# Or let HPA handle it automatically
```

### Debug Application
```bash
cd k8
make shell-backend
# Inside pod:
env | grep API
ps aux
wget -O- http://localhost:8080/health
```

### Check Performance
```bash
kubectl top pods -n godemode
kubectl get hpa -n godemode
kubectl describe hpa godemode-backend-hpa -n godemode
```

## Tips

1. **Always check context**: `kubectl config current-context`
2. **Use make help**: See all available commands
3. **Check logs first**: `make logs-backend` before debugging
4. **Let HPA work**: Don't scale manually unless needed
5. **Tag your builds**: `make build TAG=v1.2.3` for versions
6. **Test health**: `make test-health` after deploy
7. **Watch events**: `kubectl get events -n godemode` for issues
8. **Use quick-deploy**: For fast iterations during development

## Documentation

- `README.md` - Overview & architecture
- `DEPLOYMENT_GUIDE.md` - Detailed instructions
- `SUMMARY.md` - Complete setup summary
- `QUICK_REFERENCE.md` - This file
- `secrets/README.md` - Secrets management

---

**Quick Help**: Run `make help` in the `k8/` directory
