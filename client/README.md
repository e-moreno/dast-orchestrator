# DAST Orchestrator Client Scripts

This directory contains Python client scripts for interacting with the DAST Orchestrator API.

## Scripts

### `client.py` - Full Scan Client
Main client for running complete DAST scans with progress monitoring.

**Usage:**
```bash
export DAST_HMAC_SECRET="your-hmac-secret"
export DAST_API_TARGET="https://your-app.com" 
export DAST_API_URL="https://your-dast-api.com"
python client.py
```

### `client.py reload` - Reload Configuration
Use the main client with `reload` command to reload API configuration and reconnect to ZAP scanner.

**Usage:**
```bash
export DAST_HMAC_SECRET="your-hmac-secret"
export DAST_API_URL="https://your-dast-api.com"  # optional
python client.py reload
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DAST_HMAC_SECRET` | ✅ Yes | - | HMAC secret for API authentication |
| `DAST_API_URL` | No | `http://localhost:8080` | DAST API endpoint |
| `DAST_API_TARGET` | No | `https://ginandjuice.shop/` | Target URL to scan (client.py only) |
| `DAST_TARGET_APP` | No | `dast-api` | Application name (client.py only) |
| `DAST_BUILD_ID` | No | Auto-generated UUID | Build identifier (client.py only) |

## Examples

### Local Development
```bash
# Start local stack
cd ../local
docker-compose up

# Reload configuration (separate terminal)
cd ../client
export DAST_HMAC_SECRET="736ffa5e4064da13711d075ed6b71069"
python client.py reload
```

### Production
```bash
# Set environment variables
export DAST_HMAC_SECRET="your-production-secret"
export DAST_API_URL="https://dast-api.yourcompany.com"

# Reload configuration
python client.py reload

# Run a scan
export DAST_API_TARGET="https://your-app.com"
python client.py
```

## Requirements

Install dependencies:
```bash
pip install -r requirements.txt
```

## Output Examples

### Successful Reload
```
🔄 Reloading DAST configuration...

Api status:
{
    "api": "ok",
    "zap": "ok",
    "dbro": "ok",
    "dbrw": "ok"
}

✅ Reload successful!
{
    "status": "reloaded",
    "zap": "http://localhost:8090"
}
```

### Error Example
```
❌ Reload failed: 401
Authentication failed
```
