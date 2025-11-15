#!/bin/bash

# Godemode Kubernetes Deployment Script
set -e

echo "🚀 Deploying Godemode to Kubernetes..."
echo "======================================"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check current context
CONTEXT=$(kubectl config current-context)
echo "📍 Current kubectl context: ${CONTEXT}"
read -p "Continue with this context? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled"
    exit 1
fi

# Apply namespace first
echo ""
echo "📦 Creating namespace..."
kubectl apply -f namespace.yaml

# Apply ConfigMap
echo ""
echo "⚙️  Applying ConfigMap..."
kubectl apply -f configmap.yaml

# Check if secrets exist and apply them
if [ -d "secrets" ] && [ "$(ls -A secrets/*.yaml 2>/dev/null)" ]; then
    echo ""
    echo "🔐 Applying secrets..."
    kubectl apply -f secrets/
else
    echo ""
    echo "⚠️  No secrets found in k8/secrets/ directory"
    echo "   You may need to create secrets manually if required"
fi

# Apply services first (before deployments)
echo ""
echo "🌐 Applying services..."
kubectl apply -f backend-service.yaml
kubectl apply -f frontend-service.yaml

# Apply deployments
echo ""
echo "🚢 Applying deployments..."
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml

# Wait for deployments to be ready
echo ""
echo "⏳ Waiting for backend deployment to be ready..."
kubectl rollout status deployment/godemode-backend -n godemode --timeout=300s

echo ""
echo "⏳ Waiting for frontend deployment to be ready..."
kubectl rollout status deployment/godemode-frontend -n godemode --timeout=300s

# Apply HPAs (Horizontal Pod Autoscalers)
echo ""
echo "📈 Applying HPAs..."
kubectl apply -f backend-hpa.yaml
kubectl apply -f frontend-hpa.yaml

# Apply ingress
echo ""
echo "🔗 Applying ingress..."
kubectl apply -f ingress.yaml

echo ""
echo "✅ Deployment complete!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Deployment Status:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Pods:"
kubectl get pods -n godemode
echo ""
echo "Services:"
kubectl get services -n godemode
echo ""
echo "Ingress:"
kubectl get ingress -n godemode
echo ""
echo "HPAs:"
kubectl get hpa -n godemode
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 Access your application at:"
echo "   https://godemode.scalebase.io"
echo ""
echo "📝 Useful commands:"
echo "   View backend logs:  kubectl logs -f deployment/godemode-backend -n godemode"
echo "   View frontend logs: kubectl logs -f deployment/godemode-frontend -n godemode"
echo "   Get pods:           kubectl get pods -n godemode"
echo "   Describe pod:       kubectl describe pod <pod-name> -n godemode"
echo "   Delete deployment:  kubectl delete namespace godemode"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
