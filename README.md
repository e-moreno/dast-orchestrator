# 🛡️ DAST Orchestrator

A lightweight Go API that orchestrates dynamic application security testing (DAST) using **OWASP ZAP**. Designed for seamless CI/CD integration via secure webhooks.

## 🚀 Quick Start

### Docker Compose (Recommended for Testing)
```bash
cd local
docker-compose up --build
```
Access: `http://localhost:8080/ping`

### Kubernetes (Production)
```bash
# Update database config in k8s/local/configmap.yaml
cd k8s/local/
./deploy.sh
```
Access: `http://localhost:30080/ping`

## 📡 API Endpoints

### Health Check
```bash
GET /ping
# Returns: {"api":"ok","zap":"ok","dbro":"ok","dbrw":"ok"}
```

### Start Scan
```bash
POST /scan
Headers: X-Signature: <HMAC-SHA256>
Body: {
  "target": "https://example.com",
  "build_id": "abc123",
  "application": "my-app",
  "source": "ci-cd"
}
```

### Check Status
```bash
POST /status  
Headers: X-Signature: <HMAC-SHA256>
Body: {"ScanID": "abc123"}
# Returns: {"status": "running", "progress": 67}
```

### Reload Configuration
```bash
POST /reload
Headers: X-Signature: <HMAC-SHA256>
Body: {"action": "reload"}
# Returns: {"status": "reloaded", "zap": "http://localhost:8090"}
```

## 🏗️ Architecture

```text
┌─────────────────────────────────┐
│        Kubernetes Pod           │
│  ┌─────────────┐ ┌─────────────┐│
│  │  DAST API   │ │ ZAP Scanner ││
│  │   :8080     │◄┤   :8090     ││
│  └─────────────┘ └─────────────┘│
│         │                       │
└─────────┼───────────────────────┘
          │
 ┌────────▼────────┐
 │ Managed Database│
 │ (AWS RDS, DO,   │
 │  Google Cloud)  │
 └─────────────────┘
```

## ⚙️ Configuration

### 🔐 Secure Configuration (Kubernetes)

**Sensitive data** goes in `k8s/secrets.yaml`:
```bash
# Database passwords, API keys, HMAC secrets
DB_RO_PASSWORD: <base64-encoded-password>
DB_RW_PASSWORD: <base64-encoded-password>  
HMAC_SECRET: <base64-encoded-hex-secret>
ZAP_KEY: <base64-encoded-api-key>
```

**Non-sensitive config** goes in `k8s/configmap.yaml`:
```bash  
# Database connection details (no passwords)
DB_RO_HOST: "your-db-host.amazonaws.com"
DB_RO_PORT: "3306"
DB_RO_USERNAME: "dast_readonly"
# ZAP configuration
ZAP_HOST: "localhost" 
ZAP_PORT: "8090"
```

### 📝 Local Development

Set these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ZAP_HOST` | `localhost` | ZAP scanner hostname |
| `ZAP_PORT` | `8090` | ZAP scanner port |
| `ZAP_KEY` | `change-me-9203935709` | ZAP API key |
| `HMAC_SECRET` | Auto-generated | Webhook validation key |
| `DB_RO` | See configmap | Read-only database config |
| `DB_RW` | See configmap | Read-write database config |

## 🔍 Vulnerability Scoring

- **Pass**: Total vulnerability score < 8
- **Fail**: Total vulnerability score ≥ 8
- **Critical vulnerabilities** (SQL Injection, RCE): Score 8-20
- **High/Medium vulnerabilities**: Score 0-4

## ✅ CI/CD Integration

```yaml
# GitHub Actions Example
- name: DAST Scan
  env:
    DAST_HMAC_SECRET: ${{ secrets.DAST_HMAC_SECRET }}
    DAST_API_TARGET: https://my-app.com
    DAST_API_URL: https://my-dast-api.com
  run: python client/client.py
```

## 📚 Documentation

- [**Detailed Architecture**](docs/ARCHITECTURE.md) - Multi-container pod design
- [**Kubernetes Deployment**](docs/KUBERNETES.md) - Production deployment guide  
- [**API Reference**](docs/API.md) - Complete endpoint documentation
- [**Vulnerability Detection**](docs/VULNERABILITIES.md) - 134+ vulnerability types
- [**Development Guide**](docs/DEVELOPMENT.md) - Local development & testing
- [**Troubleshooting**](docs/TROUBLESHOOTING.md) - Common issues & solutions

## 🛠️ Requirements

- **Runtime**: Go 1.23+, OWASP ZAP 2.14+
- **Local Dev**: Docker, Docker Compose
- **Production**: Kubernetes cluster, Managed database (MySQL)

## 📄 License

MIT © 2025 - Built for secure software development.