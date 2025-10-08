# 🔒 Kubernetes Secrets Management

## Setup Instructions

### 1. Create Real Secrets File
```bash
# Copy template and create real secrets file
cp secrets.yaml.template secrets.yaml

# ⚠️ NEVER commit secrets.yaml with real values!
# It's already in .gitignore for protection
```

### 2. Generate Secure Values

#### Database Password
```bash
# Generate strong password
DB_PASSWORD=$(openssl rand -base64 32)
echo "Database password: $DB_PASSWORD"

# Encode for Kubernetes
echo -n "$DB_PASSWORD" | base64
```

#### HMAC Secret (Webhook Authentication)
```bash
# Generate 64-character hex string
HMAC_SECRET=$(openssl rand -hex 32)
echo "HMAC Secret: $HMAC_SECRET"

# Encode for Kubernetes  
echo -n "$HMAC_SECRET" | base64
```

#### ZAP API Key
```bash
# Generate API key
ZAP_KEY=$(openssl rand -base64 24)
echo "ZAP API Key: $ZAP_KEY"

# Encode for Kubernetes
echo -n "$ZAP_KEY" | base64
```

### 3. Update secrets.yaml
Replace the placeholder base64 values with your generated ones.

### 4. Deploy Secrets
```bash
kubectl apply -f secrets.yaml
```

## 🔒 Security Best Practices

### ✅ DO:
- Use `secrets.yaml.template` for version control
- Generate cryptographically secure random values
- Use different secrets for each environment
- Rotate secrets regularly
- Use Kubernetes RBAC to limit secret access

### ❌ DON'T:
- Commit `secrets.yaml` with real values
- Share secrets in chat/email
- Use weak or predictable secrets
- Hardcode secrets in application code
- Log secret values

## 🔧 Environment-Specific Secrets

### Local Development
- Use placeholder values from template
- Or use local secret generation for testing

### Production
- Use strong, unique secrets
- Consider using external secret management (Vault, AWS Secrets Manager)
- Enable audit logging for secret access

## 🔄 Secret Rotation

```bash
# Update secret values
kubectl patch secret dast-secrets -n dast-orchestrator -p '{"data":{"HMAC_SECRET":"'$(echo -n "new-secret" | base64)'"}}'

# Restart deployment to pick up changes
kubectl rollout restart deployment/dast-orchestrator -n dast-orchestrator
```

## 🚨 If Secrets Are Compromised

1. **Immediately rotate** all affected secrets
2. **Update** secrets.yaml with new values
3. **Redeploy** the application
4. **Audit** access logs for unauthorized usage
5. **Investigate** how the compromise occurred
