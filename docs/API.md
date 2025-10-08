# 📡 API Reference

## Authentication

All endpoints (except `/ping`) require HMAC-SHA256 authentication:

```bash
# Generate signature
signature=$(echo -n "$body" | openssl dgst -sha256 -hmac "$HMAC_SECRET" -binary | base64)

# Add header
X-Signature: $signature
```

## Endpoints

### Health Check
```bash
GET /ping
```

**Response:**
```json
{
  "api": "ok",
  "zap": "ok", 
  "dbro": "ok",
  "dbrw": "ok"
}
```

### Reload Configuration
```bash  
POST /reload
Content-Type: application/json
X-Signature: <HMAC-SHA256>
```

**Body:**
```json
{
  "action": "reload"
}
```

**Response:**
```json
{
  "status": "reloaded",
  "zap": "http://localhost:8090"
}
```

**Error Response:**
```json
{
  "status": "failed",
  "reason": "invalid JSON body"
}
```

### Start Scan
```bash
POST /scan
Content-Type: application/json
X-Signature: <HMAC-SHA256>
```

**Body:**
```json
{
  "target": "https://example.com",
  "build_id": "abc123", 
  "application": "my-app",
  "source": "ci-cd"
}
```

**Response:**
```json
{
  "scanID": "42",
  "status": "started"
}
```

### Check Scan Status
```bash
POST /status
Content-Type: application/json  
X-Signature: <HMAC-SHA256>
```

**Body:**
```json
{
  "ScanID": "abc123"
}
```

**Response (In Progress):**
```json
{
  "status": "running",
  "progress": 67
}
```

**Response (Complete):**
```json
{
  "status": "passed",
  "progress": 100,
  "vulnerabilities": [
    {
      "name": "Cross Site Scripting (Reflected)",
      "severity": "medium",
      "score": 0
    }
  ],
  "total_score": 0
}
```

## HMAC Authentication Example

### Python
```python
import hmac
import hashlib
import base64
import json

def generate_signature(payload, secret):
    message = json.dumps(payload).encode('utf-8')
    signature = hmac.new(
        secret.encode('utf-8'),
        message, 
        hashlib.sha256
    ).digest()
    return base64.b64encode(signature).decode('utf-8')

# Usage
payload = {"target": "https://example.com", "build_id": "123"}
sig = generate_signature(payload, "your-hmac-secret")
headers = {"X-Signature": sig}
```

### Bash
```bash
#!/bin/bash
HMAC_SECRET="your-hmac-secret"
BODY='{"target":"https://example.com","build_id":"123"}'

SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$HMAC_SECRET" -binary | base64)

curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Signature: $SIGNATURE" \
  -d "$BODY" \
  https://your-dast-api.com/scan
```

### Reload Example
```bash
#!/bin/bash
HMAC_SECRET="your-hmac-secret"
BODY='{"action":"reload"}'

SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$HMAC_SECRET" -binary | base64)

curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Signature: $SIGNATURE" \
  -d "$BODY" \
  https://your-dast-api.com/reload
```

## Error Responses

### Invalid Signature
```json
{
  "error": "Invalid signature"
}
```

### ZAP Not Connected
```json
{
  "status": "failed",
  "reason": "not connected to zap scanner instance"
}
```

### Invalid Request Body
```json
{
  "status": "failed", 
  "reason": "couldn't parse scan information from body"
}
```
