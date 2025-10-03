#!/bin/bash

# DAST Orchestrator Local Kubernetes Deployment Script

set -e

echo "🏠 Deploying DAST Orchestrator to Local Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Detect local Kubernetes environment
KUBE_CONTEXT=$(kubectl config current-context)
echo "📋 Current kubectl context: $KUBE_CONTEXT"

# Check for common local Kubernetes setups
if [[ $KUBE_CONTEXT == *"minikube"* ]]; then
    KUBE_ENV="minikube"
    echo "🚀 Detected Minikube environment"
elif [[ $KUBE_CONTEXT == *"kind"* ]]; then
    KUBE_ENV="kind"
    echo "🐳 Detected kind (Kubernetes in Docker) environment"
elif [[ $KUBE_CONTEXT == *"docker-desktop"* ]]; then
    KUBE_ENV="docker-desktop"
    echo "🐳 Detected Docker Desktop Kubernetes environment"
elif [[ $KUBE_CONTEXT == *"k3s"* ]] || [[ $KUBE_CONTEXT == *"k3d"* ]]; then
    KUBE_ENV="k3s"
    echo "🐄 Detected K3s/K3d environment"
else
    echo "⚠️  Unknown Kubernetes environment: $KUBE_CONTEXT"
    echo "🤔 Proceeding with generic local configuration..."
    KUBE_ENV="generic"
fi

# Setup local container registry if needed
echo "🗂️  Setting up local container registry..."
if [[ $KUBE_ENV == "minikube" ]]; then
    echo "📦 Using Minikube's built-in registry..."
    minikube addons enable registry
    kubectl port-forward --namespace kube-system service/registry 5000:80 &
    REGISTRY_PID=$!
    echo "🔗 Registry forwarded to localhost:5000 (PID: $REGISTRY_PID)"
    sleep 5
elif [[ $KUBE_ENV == "kind" ]]; then
    echo "📦 Setting up kind registry..."
    # Check if registry container exists
    if ! docker ps | grep -q "kind-registry"; then
        echo "🆕 Creating kind registry..."
        docker run -d --restart=always -p 5000:5000 --name kind-registry registry:2
        # Connect registry to kind network
        docker network connect "kind" "kind-registry" 2>/dev/null || true
    fi
elif [[ $KUBE_ENV == "docker-desktop" ]]; then
    echo "📦 Using Docker Desktop built-in registry support..."
else
    echo "📦 Setting up local registry on port 5000..."
    if ! docker ps | grep -q "local-registry"; then
        docker run -d --restart=always -p 5000:5000 --name local-registry registry:2
    fi
fi

# Build and push images to local registry
echo "🏗️  Building Docker images..."

echo "📦 Building API image..."
docker build -t localhost:5000/dast-api:latest ../../api/
docker push localhost:5000/dast-api:latest

echo "📦 Building ZAP Scanner image..."
docker build -t localhost:5000/zap-scanner:latest ../../zap/
docker push localhost:5000/zap-scanner:latest

echo "✅ Images pushed to local registry"

# Update storage configuration based on environment
echo "⚙️  Configuring storage for $KUBE_ENV..."
if [[ $KUBE_ENV == "minikube" ]]; then
    sed -i.bak 's|/tmp/dast-mysql-data|/data/dast-mysql|g' storage.yaml
    sed -i.bak 's|- minikube|- minikube|g' storage.yaml
elif [[ $KUBE_ENV == "kind" ]]; then
    # kind uses different node names
    NODE_NAME=$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')
    sed -i.bak "s|- minikube|- $NODE_NAME|g" storage.yaml
fi

# Apply Kubernetes manifests
echo "📦 Creating namespace..."
kubectl apply -f ../namespace.yaml

echo "🔐 Creating secrets..."
kubectl apply -f ../secrets.yaml

echo "⚙️  Creating local configuration..."
kubectl apply -f configmap.yaml

echo "💾 Setting up local storage..."
kubectl apply -f storage.yaml

echo "📊 Deploying MySQL database pod..."
kubectl apply -f ../database-pod.yaml

echo "🚀 Deploying main application..."
kubectl apply -f deployment.yaml

echo "🌐 Creating services..."
kubectl apply -f service.yaml

echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/dast-orchestrator -n dast-orchestrator

echo "✅ Deployment complete!"

echo "📊 Checking pod status..."
kubectl get pods -n dast-orchestrator

echo "🌐 Service information:"
kubectl get services -n dast-orchestrator

# Setup port forwarding or provide access instructions
echo ""
echo "🎉 DAST Orchestrator is now running on Local Kubernetes!"
echo ""

if [[ $KUBE_ENV == "minikube" ]]; then
    MINIKUBE_IP=$(minikube ip)
    echo "🔗 Access via NodePort: http://$MINIKUBE_IP:30080"
    echo "🔗 Or run: minikube service dast-api-nodeport -n dast-orchestrator"
elif [[ $KUBE_ENV == "docker-desktop" ]]; then
    echo "🔗 Access via NodePort: http://localhost:30080"
else
    echo "🔗 Access via NodePort: Check your node IP and use port 30080"
    kubectl get nodes -o wide
fi

echo ""
echo "📋 Access Options:"
echo "1. NodePort: Use the URLs above"
echo "2. Port Forward: kubectl port-forward svc/dast-api-service 8080:8080 -n dast-orchestrator"
echo "3. Test health: curl http://localhost:30080/ping (or forwarded port)"
echo ""
echo "🔧 Useful local commands:"
echo "- View pods: kubectl get pods -n dast-orchestrator -w"
echo "- View logs: kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api"
echo "- Shell into pod: kubectl exec -it deployment/dast-orchestrator -n dast-orchestrator -c dast-api -- /bin/sh"
echo "- Delete everything: kubectl delete namespace dast-orchestrator"
echo ""
echo "🗂️  Registry cleanup (when done):"
if [[ $REGISTRY_PID ]]; then
    echo "- Kill registry port-forward: kill $REGISTRY_PID"
fi
echo "- Remove local registry: docker rm -f local-registry kind-registry"

# Restore modified files
if [ -f "storage.yaml.bak" ]; then
    mv storage.yaml.bak storage.yaml
    echo "📄 Restored original storage.yaml"
fi

