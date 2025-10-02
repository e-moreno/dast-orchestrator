# 🛡️ DAST Orchestrator

A lightweight, modular API in Go that orchestrates dynamic application security testing (DAST) using open source scanners like **OWASP ZAP**. Designed for seamless integration into CI/CD pipelines via secure, HMAC-authenticated webhooks.

---

## 🚀 Features

- 🧪 **Automated DAST**: Run OWASP ZAP scans against deployed applications.
- 🔐 **Secure Webhooks**: Authenticate requests via HMAC (SHA256).
- 🔁 **CI/CD Integration**: Compatible with GitHub Actions, Jenkins, GitLab CI, etc.
- 📦 **Modular Scanner Support**: Add new scanners easily via Go packages.
- 📊 **Status Reporting**: Query scan progress and results via API.
- 🐳 **Dockerized**: Easy to run locally or in cloud environments.

---

## 📦 Architecture Overview

```text
          CI/CD Pipeline (GitHub Actions / Jenkins)
                       |
                   [Webhook]
                       |
                 +-----------+
                 | Orchestrator API (Go)
                 +-----------+
                       |
                +--------------+
                | OWASP ZAP (Docker)
                +--------------+
                       |
                +----------------+
                | Target Web App |
                +----------------+
````

---

## ⚙️ How It Works

1. **CI/CD pipeline** sends a `POST` request to `/scan` with the target URL and HMAC signature.
2. **Orchestrator API** validates the request, spins up an OWASP ZAP scan, and tracks its progress.
3. **ZAP** performs a spider + active scan on the target URL.
4. Scan results are returned through `/status` or stored in a DB (optional).
5. CI/CD pipelines can use the result to pass/fail builds automatically.

---

## 📡 API Endpoints

### `POST /scan`

Initiate a new scan.

**Headers:**

* `X-Signature`: HMAC-SHA256 of the payload.

**Body:**

```json
{
  "target": "http://your-target-app.com",
  "project": "example-service"
}
```

**Response:**

```json
{
  "scan_id": "abcd1234",
  "status": "started"
}
```

---

### `GET /status?id=abcd1234`

Check the scan status.

**Response:**

```json
{
  "scan_id": "abcd1234",
  "status": "running",
  "progress": 67,
  "findings": []
}
```

---

## 🛠️ Setup

### 1. Clone the repo

```bash
git clone https://github.com/your-org/dast-orchestrator.git
cd dast-orchestrator
```

### 2. Build the API

```bash
go build -o orchestrator main.go
```

### 3. Run with Docker (Recommended)

```bash
docker-compose up
```

This spins up:

* Go API (port 8080)
* OWASP ZAP daemon (port 8090)
* MySQL (optional, for storing scan results)

### 4. Environment Variables

| Variable      | Description                                 |
| ------------- | ------------------------------------------- |
| `HMAC_SECRET` | Secret key for validating webhooks          |
| `ZAP_API_KEY` | API key for OWASP ZAP                       |
| `ZAP_URL`     | URL for ZAP daemon (e.g. `http://zap:8090`) |
| `DB_DSN`      | (Optional) DSN string for MySQL             |

---

## ✅ Example GitHub Actions Workflow

```yaml
jobs:
  dast-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger DAST Scan
        run: |
          curl -X POST http://your-api/scan \
          -H "Content-Type: application/json" \
          -H "X-Signature: ${{ secrets.SCAN_SIGNATURE }}" \
          -d '{"target":"https://your-app","project":"my-app"}'
```

---

## 📚 Adding New Scanners

Each scanner is implemented as a Go package under `/scanners/`. To add support for a new DAST tool:

1. Create a new package (e.g., `scanners/arachni`)
2. Implement the `Scanner` interface
3. Register the scanner in the orchestrator

---

## 🧪 Test Targets

We recommend using:

* [OWASP Juice Shop](https://owasp.org/www-project-juice-shop/)
* [bWAPP](http://www.itsecgames.com/)
* [DVWA](http://dvwa.co.uk/)

---

## 📄 License

MIT © 2025 – Built with ❤️ for secure software development.
