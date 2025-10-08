# 📋 Deployment Checklist

## ✅ Local vs DigitalOcean Kubernetes Sync

After fixing the local Kubernetes deployment, here are the changes synchronized to DigitalOcean:

### **🔧 Changes Applied:**

#### **1. ✅ Fixed ZAP Resource Limits (Prevents OOMKilled)**
**File**: `k8s/deployment.yaml`
```yaml
# BEFORE (would cause OOMKilled during scans)
zap-scanner:
  limits:
    memory: "1Gi" 
    cpu: "500m"

# AFTER (prevents container restarts)
zap-scanner:
  limits:
    memory: "2Gi"     # 2x increase
    cpu: "1000m"      # 2x increase
```

#### **2. ✅ Fixed HMAC Secret Format**
**Files**: `k8s/digitalocean/configmap.yaml`, `k8s/secrets.yaml`
```yaml
# BEFORE (plain text - caused validation errors)
HMAC_SECRET: "do-production-hmac-secret-2024"

# AFTER (hex format - works with API validation)
HMAC_SECRET: "a1b2c3d4e5f6789012345678901234567890abcdef123456789012345678901234"
```

#### **3. ✅ Reload Endpoint Available**
- API code includes POST /reload endpoint
- Requires rebuild and push of API image

### **🚀 Deployment Commands:**

#### **For DigitalOcean:**
```bash
cd k8s/digitalocean/
./deploy.sh
```

#### **For Generic Kubernetes:**
```bash
cd k8s/
./deploy.sh
```

### **📝 Pre-Deployment Checklist:**

#### **Before Deploying to DigitalOcean:**

1. **✅ Update Registry URLs**
   - Update image references in deployment.yaml
   - Push updated API image with reload endpoint

2. **✅ Update Database Credentials**
   - Set real database passwords in configmap.yaml
   - Update host/port for your managed database

3. **✅ Update Secrets**
   - Generate production HMAC secret (hex format)
   - Set real ZAP API key
   - Set real database passwords

4. **✅ Resource Planning**
   - Ensure cluster has enough resources:
     - Per pod: 2GB RAM + 1 CPU core for ZAP
     - Scale accordingly for multiple replicas

### **🔍 Verification Steps:**

After deployment:

1. **Check Pod Status**
   ```bash
   kubectl get pods -n dast-orchestrator
   # Should show: 2/2 Running, RESTARTS: 0
   ```

2. **Check Resource Usage**
   ```bash
   kubectl top pods -n dast-orchestrator
   # ZAP memory should stay under 2GB
   ```

3. **Test API Health**
   ```bash
   curl https://your-api-url/ping
   # Should return: {"api":"ok","zap":"ok","dbro":"ok","dbrw":"ok"}
   ```

4. **Test Reload Endpoint**
   ```bash
   # Use proper HMAC secret for your environment
   curl -X POST -H "Content-Type: application/json" \
        -H "Signature: <HMAC-SHA256>" \
        -d '{"action":"reload"}' \
        https://your-api-url/reload
   ```

5. **Run Test Scan**
   ```bash
   # Monitor for container restarts during scan
   kubectl get pods -n dast-orchestrator -w
   # Should remain stable throughout scan
   ```

### **🚨 Common Issues & Solutions:**

#### **Container Still OOMKilled:**
- Increase ZAP memory beyond 2Gi if needed
- Check target application complexity
- Monitor actual memory usage: `kubectl top pods`

#### **HMAC Authentication Fails:**
- Ensure HMAC secret is in hex format
- Verify secret matches between ConfigMap and client
- Check client uses `Signature` header (not `X-Signature`)

#### **Scan Progress Lost:**
- Check for container restarts: `kubectl get pods`
- Review pod events: `kubectl describe pod <pod-name>`
- Check ZAP logs: `kubectl logs <pod-name> -c zap-scanner`

### **📊 Resource Recommendations:**

#### **Production Workloads:**
```yaml
# For high-volume scanning
zap-scanner:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "4Gi"      # For complex applications
    cpu: "2000m"       # For faster scanning
```

#### **Development/Testing:**
```yaml
# Current configuration (sufficient for most cases)
zap-scanner:
  requests:
    memory: "512Mi"
    cpu: "200m"
  limits:
    memory: "2Gi"
    cpu: "1000m"
```

---

✅ **All local fixes have been synchronized to DigitalOcean deployment files.**
