# Kubernetes Secrets

This directory contains Kubernetes Secret manifests for sensitive configuration.

## Setup

1. Copy the example file:
   ```bash
   cp backend-secrets.example.yaml backend-secrets.yaml
   ```

2. Edit `backend-secrets.yaml` with your actual secrets:
   ```bash
   # Use your preferred editor
   vim backend-secrets.yaml
   # or
   code backend-secrets.yaml
   ```

3. Apply the secrets:
   ```bash
   kubectl apply -f backend-secrets.yaml
   ```

## Important Security Notes

- **NEVER commit actual secret files to git**
- All `.yaml` files (except examples) are gitignored automatically
- Store production secrets in a secure secrets manager (e.g., HashiCorp Vault, AWS Secrets Manager)
- For CI/CD, use environment variables or sealed secrets

## Required Secrets

### Backend Secrets

- `ANTHROPIC_API_KEY`: Your Anthropic Claude API key (required)
  - Get from: https://console.anthropic.com/

### Optional Secrets

Depending on your deployment needs, you may add:
- Database credentials
- JWT signing keys
- Encryption keys
- Third-party API keys

## Updating Secrets

To update an existing secret:

```bash
# Edit the file
vim backend-secrets.yaml

# Apply the changes
kubectl apply -f backend-secrets.yaml

# Restart pods to pick up new secrets
kubectl rollout restart deployment/godemode-backend -n godemode
```

## Using Secrets in Deployments

Secrets are referenced in `backend-deployment.yaml`:

```yaml
envFrom:
  - secretRef:
      name: godemode-backend-secrets
```

This automatically loads all keys as environment variables.

## Alternative: Create Secrets via kubectl

Instead of YAML files, you can create secrets directly:

```bash
kubectl create secret generic godemode-backend-secrets \
  --from-literal=ANTHROPIC_API_KEY=your-key-here \
  --namespace=godemode
```

## Viewing Secrets

```bash
# List secrets
kubectl get secrets -n godemode

# Describe secret (doesn't show values)
kubectl describe secret godemode-backend-secrets -n godemode

# View secret values (base64 encoded)
kubectl get secret godemode-backend-secrets -n godemode -o yaml

# Decode a specific value
kubectl get secret godemode-backend-secrets -n godemode -o jsonpath='{.data.ANTHROPIC_API_KEY}' | base64 -d
```
