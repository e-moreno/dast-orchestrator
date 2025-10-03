# 🛡️ DAST Orchestrator

A lightweight, modular API in Go that orchestrates dynamic application security testing (DAST) using open source scanners like **OWASP ZAP**. Designed for seamless integration into CI/CD pipelines via secure, HMAC-authenticated webhooks.

---

## 🚀 Features

- 🧪 **Automated DAST**: Run OWASP ZAP scans against deployed applications.
- 🔐 **Secure Webhooks**: Authenticate requests via HMAC (SHA256).
- 🔁 **CI/CD Integration**: Compatible with GitHub Actions, Jenkins, GitLab CI, etc.
- 📦 **Modular Scanner Support**: Add new scanners easily via Go packages.
- 📊 **Status Reporting**: Query scan progress and results via API.
- 🐳 **Containerized**: Docker and Kubernetes ready.
- ☸️ **Kubernetes Native**: Production-ready multi-container pod architecture.
- 🔒 **Secure by Default**: Updated base images and secure configurations.

---

## 📦 Architecture Overview

### Multi-Container Pod Architecture

```text
┌────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                      │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                 DAST Pod                            │   │
│  │  ┌─────────────┐    ┌─────────────┐                 │   │
│  │  │  DAST API   │    │ ZAP Scanner │                 │   │
│  │  │  :8080      │◄──►│  :8090      │                 │   │
│  │  └─────────────┘    └─────────────┘                 │   │
│  │         │                                           │   │
│  └─────────┼───────────────────────────────────────────┘   │
│            │                                               │
│  ┌─────────▼───────┐     OR     ┌─────────────────────┐    │
│  │  MySQL Pod      │            │   Managed Database  │    │
│  │  (Optional)     │            │   (AWS RDS, etc.)   │    │
│  └─────────────────┘            └─────────────────────┘    │
└────────────────────────────────────────────────────────────┘
```

### Traditional Docker Compose Architecture

```text
          CI/CD Pipeline (GitHub Actions / Jenkins)
                             |
                         [Webhook]
                             |
                 +-----------------------+
                 | Orchestrator API (Go) |
                 +-----------------------+
                             |
                +--------------------+
                | OWASP ZAP (Docker) |
                +--------------------+
                           |
                +----------------+
                | Target Web App |
                +----------------+
```

---

## ⚙️ How It Works

1. **CI/CD pipeline** sends a `POST` request to `/scan` with the target URL, build ID, and HMAC signature.
2. **Orchestrator API** validates the HMAC signature, creates a new scan record, and initiates a ZAP scan.
3. **ZAP performs two-phase scanning**:
   - **Spider scan**: Crawls the target application to discover all accessible endpoints and pages
   - **Active scan**: Tests each discovered endpoint for vulnerabilities using comprehensive security rules
4. **Async monitoring**: The system monitors scan progress in the background and updates the database when complete.
5. **Vulnerability scoring**: Maps detected issues to CWE codes and calculates a total risk score.
6. **Pass/Fail decision**: Scans with total score < 8 pass, scores ≥ 8 fail the build.
7. **CI/CD integration**: Pipelines can query scan status via `/status` and automatically pass/fail builds.

---

## 📡 API Endpoints

### Health Check
- **GET** `/ping` - Service health check

**Response:**
```json
{
  "api": "ok",
  "zap": "ok", 
  "dbro": "ok",
  "dbrw": "ok"
}
```

### Service Management
- **GET** `/reload/b27ddce7-cbc4-4556-927e-7ae0203cb66c` - Reload configuration and reconnect to services (production endpoint with UUID protection)

### `POST /scan`

Initiate a new scan.

**Headers:**

* `X-Signature`: HMAC-SHA256 of the payload.

**Body:**

```json
{
  "target": "http://your-target-app.com",
  "build_id": "abc123",
  "application": "example-service",
  "source": "ci-cd"
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

### `POST /status`

Check the scan status using the build ID.

**Headers:**

* `X-Signature`: HMAC-SHA256 of the payload.

**Body:**

```json
{
  "ScanID": "abc123"
}
```

**Response:**

```json
{
  "status": "running",
  "progress": 67
}
```

**Note:** There's also a deprecated `GET /scan/:id` endpoint that accepts the ZAP scan ID directly.

---

## 🔍 Vulnerability Detection & Scoring

### Supported Vulnerability Types

The system detects **134 different vulnerability types** based on CWE (Common Weakness Enumeration) classifications, including:

#### Critical Vulnerabilities (Score: 8-20)
- **SQL Injection** (CWE-89): Score 20
- **Remote Code Execution** (CWE-78): Score 20  
- **Source Code Disclosure - Git** (CWE-541): Score 20
- **Remote File Inclusion** (CWE-98): Score 4
- **Path Traversal** (CWE-22): Score 4
- **Server Side Template Injection** (CWE-74): Score 4

#### High Risk Vulnerabilities (Score: 1-4)
- **Cross Site Scripting (Persistent)** (CWE-79): Score 1
- **XSLT Injection** (CWE-91): Score 1
- **XML External Entity Attack** (CWE-611): Score 1
- **Directory Browsing** (CWE-548): Score 1

#### Medium/Low Risk (Score: 0)
- **Cross Site Scripting (Reflected)** (CWE-79): Score 0
- **HTTP Parameter Pollution** (CWE-20): Score 0
- **External Redirect** (CWE-601): Score 0

### Scoring System

- **Pass Threshold**: Total vulnerability score < 8
- **Fail Threshold**: Total vulnerability score ≥ 8
- **Scoring Logic**: Each detected vulnerability adds its score to the total
- **Build Decision**: CI/CD pipelines can automatically pass/fail based on the final score

### Scanning Process

1. **Spider Scan**: Crawls the target application to discover all accessible endpoints
2. **Active Scan**: Tests each discovered endpoint for vulnerabilities using OWASP ZAP's active scan rules
3. **Result Processing**: Maps detected issues to CWE codes and calculates total risk score
4. **Database Storage**: Stores scan results, findings, and vulnerability details for tracking
5. **Pass/Fail Decision**: Returns final status based on configurable score thresholds

---

## 🗄️ Database Schema

The system uses MySQL to store scan results and vulnerability information:

### Core Tables

#### `scans`
Tracks scan executions and their status:
```sql
CREATE TABLE scans (
    id INT PRIMARY KEY AUTO_INCREMENT,
    status VARCHAR(255),           -- 'started', 'running', 'passed', 'failed'
    build_id VARCHAR(255),         -- CI/CD build identifier
    build_source VARCHAR(255),     -- 'github', 'jenkins', etc.
    target VARCHAR(255),           -- Target URL scanned
    zap_id INT,                    -- OWASP ZAP scan ID
    asset_id INT,                  -- Reference to application/project
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

#### `vulnerabilities`
Pre-configured vulnerability types with CWE mappings:
```sql
CREATE TABLE vulnerabilities (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),             -- Human-readable vulnerability name
    cwe_id INT,                    -- CWE classification number
    severity ENUM('low','medium','high','critical'),
    score INT                      -- Risk score for pass/fail logic
);
```

#### `vulnerability_findings`
Links scan results to specific vulnerabilities found:
```sql
CREATE TABLE vulnerability_findings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    scan_id INT,                   -- References scans.id
    vulnerability_id INT,          -- References vulnerabilities.id
    details LONGTEXT               -- JSON details from ZAP scan
);
```

### Database Configuration

The system supports **dual database connections**:
- **Read-Only (RO)**: For status queries and reporting
- **Read-Write (RW)**: For scan result storage and updates

**Resilience**: The system gracefully handles database connection failures and continues operating without persistence if needed.

---

## 🛠️ Local Development Setup

### Prerequisites
- Docker and Docker Compose
- Git

### Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/your-org/dast-orchestrator.git
cd dast-orchestrator
```

2. **Run with Docker Compose (Recommended)**
```bash
cd local
docker-compose up --build
```

This will start:
- **API Server** (port 8080) - Main orchestrator service
- **OWASP ZAP Scanner** (port 8090) - Security scanning engine  
- **MySQL Database** (port 3306) - Scan results storage

### Verification

1. **Check service health:**
```bash
curl http://localhost:8080/ping
```

2. **Verify ZAP scanner:**
```bash
curl -H "X-ZAP-API-Key: change-me-9203935709" \
     http://localhost:8090/JSON/core/view/version/
```

3. **Check logs:**
```bash
docker-compose logs dast-api
```

### Configuration

The application uses secure defaults for local development:

| Service | Default Host | Port | Notes |
|---------|-------------|------|-------|
| API | `localhost` | 8080 | Main service |
| ZAP Scanner | `zap-scanner` | 8090 | Internal network |
| Database | `dast-db` | 3306 | Internal network |

### Environment Variables

Default values are optimized for local development. Override in `local/local.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `ZAP_HOST` | `zap-scanner` | ZAP scanner hostname |
| `ZAP_PORT` | `8090` | ZAP scanner port |
| `ZAP_KEY` | `change-me-9203935709` | ZAP API key |
| `HMAC_SECRET` | `736ffa5e4064da13711d075ed6b71069` | Webhook validation |
| `DB_RO` | Local MySQL config | Read-only database |
| `DB_RW` | Local MySQL config | Read-write database |

---

## ☸️ Kubernetes Production Deployment

### Multi-Container Pod Design

The Kubernetes deployment uses a **multi-container pod** architecture with these benefits:

- ✅ **Tight Coupling**: API and ZAP always deployed together
- ✅ **Shared Network**: Both containers share `localhost` network
- ✅ **Shared Storage**: Can share volumes if needed
- ✅ **Atomic Scaling**: Scale as one unit
- ✅ **Resource Efficiency**: No network overhead between API and ZAP

### Deployment Options

#### 🌊 DigitalOcean Kubernetes

**Features:**
- DigitalOcean LoadBalancer integration
- DigitalOcean Block Storage for persistence
- DigitalOcean Container Registry support
- DigitalOcean Managed Database compatibility

**Deploy:**
```bash
cd k8s/digitalocean/
./deploy.sh
```

**What it sets up:**
- Automated image build and push to DO Container Registry
- LoadBalancer with health checks
- Block Storage for database persistence
- Integration with DO Managed Database services

#### 🏠 Local Kubernetes (Testing)

**Supports:**
- Minikube
- kind (Kubernetes in Docker)
- Docker Desktop Kubernetes
- K3s/K3d

**Deploy:**
```bash
cd k8s/local/
./deploy.sh
```

**What it sets up:**
- Local container registry
- NodePort services for access
- Local storage for testing
- Reduced resource requirements

#### 🌐 Generic Kubernetes (Cloud Agnostic)

**Deploy:**
```bash
cd k8s/
./deploy.sh
```

**Customize for your cloud:**
- Update image registry URLs
- Configure LoadBalancer annotations
- Set storage classes
- Update database configurations

### Kubernetes Manifests Explained

#### 📁 **File Structure**

```text
k8s/
├── namespace.yaml          # Resource isolation
├── secrets.yaml           # Sensitive data (passwords, keys)
├── configmap.yaml         # Non-sensitive configuration  
├── deployment.yaml        # Multi-container pod definition
├── service.yaml           # Network access (ClusterIP + LoadBalancer)
├── database-pod.yaml      # Optional MySQL pod
├── hpa.yaml              # Horizontal Pod Autoscaler
└── deploy.sh             # Automated deployment script

digitalocean/
├── configmap.yaml        # DO-specific configuration
├── service.yaml          # DO LoadBalancer with annotations
├── storage.yaml          # DO Block Storage configuration
└── deploy.sh            # DO-specific deployment script

local/
├── configmap.yaml        # Local development configuration
├── deployment.yaml       # Reduced resource requirements
├── service.yaml          # NodePort for local access
├── storage.yaml          # Local storage (hostPath)
└── deploy.sh            # Local deployment with registry setup
```

#### 🔧 **Key Components**

**1. Multi-Container Deployment:**
```yaml
spec:
  containers:
  - name: zap-scanner
    image: your-registry/zap-scanner:latest
    ports:
    - containerPort: 8090
    
  - name: dast-api  
    image: your-registry/dast-api:latest
    ports:
    - containerPort: 8080
    env:
    - name: ZAP_HOST
      value: "localhost"  # Same pod communication
```

**2. Database Flexibility:**
```yaml
# Option A: Managed Database
env:
- name: DB_RO
  value: '{"host":"rds-endpoint.amazonaws.com"}'

# Option B: Pod Database  
spec:
  containers:
  - name: mysql
    volumeMounts:
    - name: mysql-storage
      mountPath: /var/lib/mysql
```

**3. Auto-Scaling:**
```yaml
spec:
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
```

### Production Benefits

#### **Reliability**
- ✅ **Multi-Pod**: 2+ replicas for high availability
- ✅ **Health Checks**: Automatic restart of failed containers
- ✅ **Rolling Updates**: Zero-downtime deployments
- ✅ **Resource Limits**: Prevents resource exhaustion

#### **Scalability**  
- ✅ **Horizontal Auto-scaling**: Automatic scaling based on load
- ✅ **Load Distribution**: Traffic spread across multiple pods
- ✅ **Resource Optimization**: CPU/Memory requests and limits

#### **Security**
- ✅ **Secrets Management**: Encrypted secret storage
- ✅ **Network Policies**: Can isolate network traffic
- ✅ **Non-Root**: Containers run as non-root users
- ✅ **Namespace Isolation**: Separate from other applications

#### **Operations**
- ✅ **Monitoring**: Built-in Kubernetes metrics
- ✅ **Logging**: Centralized log collection
- ✅ **Updates**: Easy configuration updates via ConfigMaps
- ✅ **Rollback**: Easy rollback to previous versions

### Scaling & Load Distribution

```text
Load Balancer
     │
┌────▼────┐    ┌─────────┐    ┌─────────┐
│ Pod 1   │    │ Pod 2   │    │ Pod N   │
│ API+ZAP │    │ API+ZAP │    │ API+ZAP │
└─────────┘    └─────────┘    └─────────┘
     │              │              │
     └──────────────┼──────────────┘
                    │
            ┌───────▼──────┐
            │   Database   │
            └──────────────┘
```

### Pod Communication

```text
┌─────────────────────────────────┐
│          DAST Pod               │
│                                 │
│  API Container  ──localhost──►  │ ZAP Container
│  (port 8080)      (no network   │ (port 8090)
│                    overhead)    │
└─────────────────────────────────┘
```

---

## 🔍 Troubleshooting

### Docker Compose Issues

**Service won't start:**
```bash
# Check logs
docker-compose logs dast-api

# Verify all services are healthy
docker-compose ps
```

**Database connection issues:**
```bash
# Test database connectivity
docker-compose exec dast-db mysql -u user -ppassword -e "SELECT 1"
```

**ZAP scanner not responding:**
```bash
# Check ZAP health
curl -H "X-ZAP-API-Key: change-me-9203935709" \
     http://localhost:8090/JSON/core/view/version/
```

**API returns empty response:**
- Ensure all services show "healthy" status
- Check `docker-compose logs` for startup errors
- Verify port 8080 is not in use

### Kubernetes Issues

**Pods not starting:**
```bash
# Check pod status
kubectl get pods -n dast-orchestrator

# Check pod events
kubectl describe pod <pod-name> -n dast-orchestrator

# Check logs
kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api
```

**Image pull issues:**
```bash
# Check if images exist in registry
docker images | grep dast

# For local clusters, ensure images are pushed to local registry
docker push localhost:5000/dast-api:latest
```

**Service not accessible:**
```bash
# Check services
kubectl get svc -n dast-orchestrator

# Port forward for testing
kubectl port-forward svc/dast-api-service 8080:8080 -n dast-orchestrator

# Test health
curl http://localhost:8080/ping
```

**Database connection issues:**
```bash
# Check database pod
kubectl get pods -n dast-orchestrator | grep mysql

# Check database logs
kubectl logs -f deployment/dast-mysql -n dast-orchestrator

# Test database connection
kubectl exec -it deployment/dast-mysql -n dast-orchestrator -- mysql -u dast_user -p
```

### Service Dependencies

The docker-compose setup ensures proper startup order:
1. **Database** starts and becomes healthy
2. **ZAP Scanner** starts and becomes healthy  
3. **API** starts only after both dependencies are ready

The Kubernetes setup uses:
- **Health checks** for readiness and liveness
- **Init containers** for database initialization
- **Dependency ordering** via `depends_on` conditions

---

## 🔧 Development

### Security Improvements (Recent Updates)

- ✅ **Updated base images**: Go 1.23, Alpine 3.20 (latest security patches)
- ✅ **Secure defaults**: Local development uses internal Docker network
- ✅ **Health checks**: Proper service dependency management
- ✅ **Connection timeouts**: Graceful handling of external service failures
- ✅ **Improved logging**: Detailed startup and error reporting
- ✅ **Multi-environment configs**: Separate configs for local, DO, and generic K8s

### Error Handling & Resilience

The system is designed to handle various failure scenarios gracefully:

- **Database Connection Failures**: System continues operating without persistence when DB is unavailable
- **ZAP Scanner Disconnection**: API returns appropriate error messages and can reconnect via `/reload` endpoint
- **Network Timeouts**: Configurable timeouts for scan operations and external service calls
- **Concurrent Scans**: Session management prevents scan interference using build IDs
- **Partial Failures**: Individual scan failures don't affect other running scans

### Build from Source

```bash
cd api
go mod download
go build -o main ./cmd
./main
```

### Running Tests

```bash
cd api
go test ./...
```

### Building Images

**For Docker Compose:**
```bash
docker-compose build
```

**For Kubernetes:**
```bash
# Local registry
docker build -t localhost:5000/dast-api:latest api/
docker push localhost:5000/dast-api:latest

# Remote registry
docker build -t your-registry/dast-api:latest api/
docker push your-registry/dast-api:latest
```

---

## ✅ Example GitHub Actions Workflow

```yaml
name: DAST-Scanner

on:
  push:
    branches:
      - main
    paths-ignore:
      - "**.md"
      - "infrastructure/"
      - "deploy/**"
      - ".github/**"
      - "local/"
      - ".gitignore"

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: checkout repo content
        uses: actions/checkout@v2
      - name: setup python
        uses: actions/setup-python@v2
        with:
          python-version: 3.8
      - name: Install dependencies
        run: |
            python -m pip install --upgrade pip
            if [ -f client/requirements.txt ]; then pip install -r client/requirements.txt; fi
      - name: Run scanner
        env:
          DAST_HMAC_SECRET: ${{ secrets.DAST_HMAC_SECRET }}
          DAST_API_TARGET: https://ginandjuice.shop/
          DAST_API_URL: <PUBLIC_API_URL>
          DAST_TARGET_APP: dast-api
          DAST_BUILD_ID: ${{ github.sha }}
        run: |
          python client/client.py
```

### Python Client Details

The included Python client (`client/client.py`) provides a complete example of DAST integration:

1. **Health Check**: Verifies API and ZAP scanner availability before starting
2. **HMAC Authentication**: Generates proper signatures for API requests  
3. **Scan Initiation**: Sends target URL and build metadata to start scanning
4. **Progress Monitoring**: Polls scan status with a progress bar until completion
5. **Result Processing**: Returns final pass/fail status for CI/CD decision making

**Environment Variables:**
- `DAST_API_TARGET`: Target URL to scan (default: https://ginandjuice.shop/)
- `DAST_API_URL`: DAST API endpoint (default: http://localhost:8080)
- `DAST_HMAC_SECRET`: HMAC secret for request authentication
- `DAST_TARGET_APP`: Application name for tracking
- `DAST_BUILD_ID`: Build identifier (default: auto-generated UUID)

---

## 🧪 Test Targets

We recommend using:

* [OWASP Juice Shop](https://owasp.org/www-project-juice-shop/)
* [bWAPP](http://www.itsecgames.com/)
* [DVWA](http://dvwa.co.uk/)

---

## 📚 Adding New Scanners

Each scanner is implemented as a Go package under `pkg/`. To add support for a new DAST tool:

1. Create a new package (e.g., `pkg/arachniScanner`)
2. Implement the `ScannerService` interface
3. Register the scanner in the controller

---

## 📄 License

MIT © 2025 – Built with ❤️ for secure software development.
