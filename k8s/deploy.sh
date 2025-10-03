#!/bin/bash

# DAST Orchestrator Kubernetes Deployment Script

set -e

echo "🚀 Deploying DAST Orchestrator to Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster. Please configure kubectl."
    exit 1
fi

echo "📋 Current cluster info:"
kubectl cluster-info

# Confirm deployment
read -p "🤔 Deploy to this cluster? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Deployment cancelled."
    exit 1
fi

# Build and push images (you'll need to modify these)
echo "🏗️  Building Docker images..."
echo "📝 NOTE: Update the image registry URLs in deployment.yaml first!"

# Example build commands (customize for your registry)
# docker build -t your-registry/dast-api:latest ../api/
# docker push your-registry/dast-api:latest
# docker build -t your-registry/zap-scanner:latest ../zap/
# docker push your-registry/zap-scanner:latest

# Apply Kubernetes manifests
echo "📦 Creating namespace..."
kubectl apply -f namespace.yaml

echo "🔐 Creating secrets..."
kubectl apply -f secrets.yaml

echo "⚙️  Creating config maps..."
kubectl apply -f configmap.yaml

# Choose database option
echo "💾 Database Options:"
echo "1) Use managed database service (AWS RDS, Google Cloud SQL, etc.)"
echo "2) Deploy MySQL pod in cluster"
read -p "Choose option (1 or 2): " -n 1 -r
echo

if [[ $REPLY == "2" ]]; then
    echo "📊 Deploying MySQL database pod..."
    kubectl apply -f database-pod.yaml
    
    # Update ConfigMap for pod-based database
    kubectl patch configmap dast-config -n dast-orchestrator --patch='
    data:
      DB_RO: |
        {
          "username": "dast_user",
          "password": "PLACEHOLDER",
          "engine": "mysql", 
          "host": "dast-mysql-service.dast-orchestrator.svc.cluster.local",
          "port": 3306,
          "dbClusterIdentifier": "dastapi"
        }
      DB_RW: |
        {
          "username": "dast_user", 
          "password": "PLACEHOLDER",
          "engine": "mysql",
          "host": "dast-mysql-service.dast-orchestrator.svc.cluster.local", 
          "port": 3306,
          "dbClusterIdentifier": "dastapi"
        }
    '
    echo "✅ MySQL pod configuration applied"
else
    echo "📝 Using managed database service (make sure to update configmap.yaml with your DB details)"
fi

echo "🚀 Deploying main application..."
kubectl apply -f deployment.yaml

echo "🌐 Creating services..."
kubectl apply -f service.yaml

echo "📈 Setting up autoscaling..."
kubectl apply -f hpa.yaml

echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/dast-orchestrator -n dast-orchestrator

echo "✅ Deployment complete!"

echo "📊 Checking pod status..."
kubectl get pods -n dast-orchestrator

echo "🌐 Service information:"
kubectl get services -n dast-orchestrator

echo "📈 HPA status:"
kubectl get hpa -n dast-orchestrator

echo ""
echo "🎉 DAST Orchestrator is now running on Kubernetes!"
echo ""
echo "📋 Next steps:"
echo "1. Get external IP: kubectl get svc dast-api-loadbalancer -n dast-orchestrator"
echo "2. Test health: curl http://<EXTERNAL-IP>/ping"
echo "3. View logs: kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api"
echo "4. Scale manually: kubectl scale deployment dast-orchestrator --replicas=5 -n dast-orchestrator"
echo ""
echo "🔧 Useful commands:"
echo "- Watch pods: kubectl get pods -n dast-orchestrator -w"
echo "- Port forward: kubectl port-forward svc/dast-api-service 8080:8080 -n dast-orchestrator"
echo "- Delete everything: kubectl delete namespace dast-orchestrator"
