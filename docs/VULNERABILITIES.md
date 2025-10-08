# 🔍 Vulnerability Detection & Scoring

## Scoring System

- **Pass Threshold**: Total vulnerability score < 8
- **Fail Threshold**: Total vulnerability score ≥ 8  
- **Build Decision**: CI/CD pipelines automatically pass/fail based on final score

## Vulnerability Types (134+ Supported)

### Critical Vulnerabilities (Score: 8-20)
- **SQL Injection** (CWE-89): Score 20
- **Remote Code Execution** (CWE-78): Score 20
- **Source Code Disclosure - Git** (CWE-541): Score 20
- **Remote File Inclusion** (CWE-98): Score 4
- **Path Traversal** (CWE-22): Score 4
- **Server Side Template Injection** (CWE-74): Score 4

### High Risk Vulnerabilities (Score: 1-4)
- **Cross Site Scripting (Persistent)** (CWE-79): Score 1
- **XSLT Injection** (CWE-91): Score 1
- **XML External Entity Attack** (CWE-611): Score 1
- **Directory Browsing** (CWE-548): Score 1

### Medium/Low Risk (Score: 0)
- **Cross Site Scripting (Reflected)** (CWE-79): Score 0
- **HTTP Parameter Pollution** (CWE-20): Score 0
- **External Redirect** (CWE-601): Score 0

## Scanning Process

1. **Spider Scan**: Crawls target application to discover endpoints
2. **Active Scan**: Tests each endpoint for vulnerabilities using OWASP ZAP rules
3. **Result Processing**: Maps detected issues to CWE codes and calculates total score
4. **Database Storage**: Stores findings and vulnerability details
5. **Pass/Fail Decision**: Returns status based on configurable score thresholds

## Database Schema

### `vulnerabilities` Table
```sql
CREATE TABLE vulnerabilities (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),             -- Human-readable name
    cwe_id INT,                    -- CWE classification
    severity ENUM('low','medium','high','critical'),
    score INT                      -- Risk score
);
```

### `vulnerability_findings` Table  
```sql
CREATE TABLE vulnerability_findings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    scan_id INT,                   -- References scans.id
    vulnerability_id INT,          -- References vulnerabilities.id
    details LONGTEXT               -- JSON details from ZAP
);
```
