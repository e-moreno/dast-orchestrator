# ☸️ Kubernetes Deployment Guide

## Deployment Options

### 🏠 Local Kubernetes
**Supports**: Minikube, kind, Docker Desktop, K3s/K3d, Rancher Desktop

```bash
cd k8s/local/
# Update database config in configmap.yaml
./deploy.sh
```

**Access**: `http://localhost:30080` or port-forward: `kubectl port-forward svc/dast-api-service 8080:8080 -n dast-orchestrator`

### 🌊 DigitalOcean Kubernetes
```bash
cd k8s/digitalocean/
./deploy.sh
```

### 🌐 Generic Cloud
```bash
cd k8s/
# Update image registry and storage class
./deploy.sh
```

## Configuration

### Required: Update Database Connection
Edit `k8s/local/configmap.yaml`:
```yaml
DB_RO: |
  {
    "username": "your_db_user",
    "password": "your_db_password", 
    "engine": "mysql",
    "host": "your-managed-database-host.com",
    "port": 3306,
    "dbClusterIdentifier": "dastapi"
  }
```

## Architecture Benefits

- ✅ **Multi-container pod**: API + ZAP share localhost network
- ✅ **Auto-scaling**: HPA based on CPU/memory
- ✅ **Health checks**: Automatic restart on failure
- ✅ **Zero-downtime**: Rolling updates
- ✅ **Resource limits**: Prevents resource exhaustion

## Verification

```bash
# Check deployment status
kubectl get pods -n dast-orchestrator

# Test health
curl http://localhost:30080/ping

# View logs  
kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api
```

## Cleanup

```bash
kubectl delete namespace dast-orchestrator
```
