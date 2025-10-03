#!/bin/bash

# DAST Orchestrator DigitalOcean Kubernetes Deployment Script

set -e

echo "🌊 Deploying DAST Orchestrator to DigitalOcean Kubernetes..."

# Check if doctl is available
if ! command -v doctl &> /dev/null; then
    echo "❌ doctl not found. Please install DigitalOcean CLI first."
    echo "📖 See: https://docs.digitalocean.com/reference/doctl/how-to/install/"
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Check DigitalOcean authentication
if ! doctl auth list &> /dev/null; then
    echo "❌ Not authenticated with DigitalOcean. Please run: doctl auth init"
    exit 1
fi

echo "📋 Available DigitalOcean Kubernetes clusters:"
doctl kubernetes cluster list

# Get cluster context
read -p "🤔 Enter your cluster name: " CLUSTER_NAME
if [ -z "$CLUSTER_NAME" ]; then
    echo "❌ Cluster name required."
    exit 1
fi

echo "🔗 Setting up kubectl context for cluster: $CLUSTER_NAME"
doctl kubernetes cluster kubeconfig save $CLUSTER_NAME

echo "📋 Current cluster info:"
kubectl cluster-info

# Confirm deployment
read -p "🤔 Deploy to DigitalOcean cluster '$CLUSTER_NAME'? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Deployment cancelled."
    exit 1
fi

# Create DigitalOcean Container Registry if not exists
echo "📦 Setting up DigitalOcean Container Registry..."
read -p "🤔 Enter your registry name (e.g., my-registry): " REGISTRY_NAME
if [ -z "$REGISTRY_NAME" ]; then
    echo "❌ Registry name required."
    exit 1
fi

# Check if registry exists, create if not
if ! doctl registry get $REGISTRY_NAME &> /dev/null; then
    echo "🆕 Creating registry: $REGISTRY_NAME"
    doctl registry create $REGISTRY_NAME --region nyc3
fi

echo "🔐 Configuring Docker for DigitalOcean Registry..."
doctl registry login

# Build and push images
echo "🏗️  Building and pushing Docker images..."
REGISTRY_URL="registry.digitalocean.com/$REGISTRY_NAME"

echo "📦 Building API image..."
docker build -t $REGISTRY_URL/dast-api:latest ../../api/
docker push $REGISTRY_URL/dast-api:latest

echo "📦 Building ZAP Scanner image..."
docker build -t $REGISTRY_URL/zap-scanner:latest ../../zap/
docker push $REGISTRY_URL/zap-scanner:latest

echo "✅ Images pushed to DigitalOcean Container Registry"

# Update deployment with correct image URLs
echo "⚙️  Updating deployment configuration..."
sed -i.bak "s|your-registry/dast-api:latest|$REGISTRY_URL/dast-api:latest|g" ../deployment.yaml
sed -i.bak "s|your-registry/zap-scanner:latest|$REGISTRY_URL/zap-scanner:latest|g" ../deployment.yaml

# Apply Kubernetes manifests
echo "📦 Creating namespace..."
kubectl apply -f ../namespace.yaml

echo "🔐 Creating secrets..."
kubectl apply -f ../secrets.yaml

echo "⚙️  Creating DigitalOcean-specific config..."
kubectl apply -f configmap.yaml

echo "💾 Setting up DigitalOcean storage..."
kubectl apply -f storage.yaml

# Database choice
echo "💾 Database Options for DigitalOcean:"
echo "1) Use DigitalOcean Managed Database (Recommended)"
echo "2) Deploy MySQL pod with DigitalOcean Block Storage"
read -p "Choose option (1 or 2): " -n 1 -r
echo

if [[ $REPLY == "2" ]]; then
    echo "📊 Deploying MySQL with DigitalOcean Block Storage..."
    kubectl apply -f ../database-pod.yaml
    
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
    echo "✅ MySQL pod with DO Block Storage configured"
else
    echo "📝 Using DigitalOcean Managed Database"
    echo "🔗 Create your managed database at: https://cloud.digitalocean.com/databases"
    echo "📋 Update configmap.yaml with your database connection details"
fi

echo "🚀 Deploying main application..."
kubectl apply -f ../deployment.yaml

echo "🌐 Creating DigitalOcean LoadBalancer service..."
kubectl apply -f service.yaml

echo "📈 Setting up autoscaling..."
kubectl apply -f ../hpa.yaml

echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/dast-orchestrator -n dast-orchestrator

echo "✅ Deployment complete!"

echo "📊 Checking pod status..."
kubectl get pods -n dast-orchestrator

echo "🌐 DigitalOcean LoadBalancer information:"
kubectl get services -n dast-orchestrator

echo "📈 HPA status:"
kubectl get hpa -n dast-orchestrator

# Get LoadBalancer IP
echo "⏳ Waiting for LoadBalancer IP..."
EXTERNAL_IP=""
while [ -z $EXTERNAL_IP ]; do
    echo "Waiting for external IP..."
    EXTERNAL_IP=$(kubectl get svc dast-api-loadbalancer -n dast-orchestrator --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
    [ -z "$EXTERNAL_IP" ] && sleep 10
done

echo ""
echo "🎉 DAST Orchestrator is now running on DigitalOcean Kubernetes!"
echo ""
echo "🔗 External IP: $EXTERNAL_IP"
echo ""
echo "📋 Next steps:"
echo "1. Test health: curl http://$EXTERNAL_IP/ping"
echo "2. Set up DNS: Point your domain to $EXTERNAL_IP"
echo "3. Configure SSL: Set up cert-manager or DO managed certificates"
echo "4. Monitor: https://cloud.digitalocean.com/kubernetes/clusters/$CLUSTER_NAME"
echo ""
echo "🔧 Useful commands:"
echo "- View in DO Console: doctl kubernetes cluster kubeconfig show $CLUSTER_NAME"
echo "- Scale: kubectl scale deployment dast-orchestrator --replicas=5 -n dast-orchestrator"
echo "- Logs: kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api"
echo "- Delete: kubectl delete namespace dast-orchestrator"

# Restore original deployment file
if [ -f "../deployment.yaml.bak" ]; then
    mv ../deployment.yaml.bak ../deployment.yaml
    echo "📄 Restored original deployment.yaml"
fi

