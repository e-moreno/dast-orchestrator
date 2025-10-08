CREATE DATABASE IF NOT EXISTS dastapi;

USE dastapi;
SET SQL_MODE='ALLOW_INVALID_DATES';

-- =====================================================
-- CREATE USERS AND PERMISSIONS
-- =====================================================

-- Create read-write user for main application operations
CREATE USER IF NOT EXISTS 'dast_user'@'%' IDENTIFIED BY 'dast_rw_password_123';

-- Create read-only user for reporting/monitoring
CREATE USER IF NOT EXISTS 'dast_readonly'@'%' IDENTIFIED BY 'dast_ro_password_456';

-- Grant full permissions to read-write user
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER 
ON dastapi.* TO 'dast_user'@'%';

-- Grant read-only permissions
GRANT SELECT ON dastapi.* TO 'dast_readonly'@'%';

-- Apply permission changes
FLUSH PRIVILEGES;

CREATE TABLE IF NOT EXISTS `scans`
(
    `id`           int PRIMARY KEY AUTO_INCREMENT,
    `status`       varchar(255),
    `build_id`     varchar(255),
    `build_source` varchar(255),
    `target`       varchar(255),
    `zap_id`       int,
    `asset_id`     int,
    `created_at`   timestamp DEFAULT CURRENT_TIMESTAMP,
    `completed_at` timestamp DEFAULT 0
);

CREATE TABLE IF NOT EXISTS `assets`
(
    `id`         int PRIMARY KEY AUTO_INCREMENT,
    `owner`      varchar(255),
    `project`    varchar(255),
    `repo`       varchar(255),
    `created_at` timestamp
);

CREATE TABLE IF NOT EXISTS `vulnerabilities`
(
    `id`         int PRIMARY KEY AUTO_INCREMENT,
    `name`       varchar(255),
    `cwe_id`     int,
    `created_at` timestamp,
    `severity`   ENUM ('low', 'medium', 'high', 'critical'),
    `score`      int
);

CREATE TABLE IF NOT EXISTS `vulnerability_findings`
(
    `id`               int PRIMARY KEY AUTO_INCREMENT,
    `scan_id`          int,
    `vulnerability_id` int,
    `details`          longtext
);

CREATE TABLE IF NOT EXISTS `configurations`
(
    `id`         int PRIMARY KEY AUTO_INCREMENT,
    `key`        varchar(255),
    `value`      varchar(255),
    `created_at` timestamp
);

ALTER TABLE `vulnerability_findings`
    ADD FOREIGN KEY (`scan_id`) REFERENCES `scans` (`id`);

ALTER TABLE `vulnerability_findings`
    ADD FOREIGN KEY (`vulnerability_id`) REFERENCES `vulnerabilities` (`id`);

ALTER TABLE `scans`
    ADD FOREIGN KEY (`asset_id`) REFERENCES `assets` (`id`);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================

-- Add indexes for better query performance
ALTER TABLE `scans` ADD INDEX idx_build_id (`build_id`);
ALTER TABLE `scans` ADD INDEX idx_status (`status`);
ALTER TABLE `scans` ADD INDEX idx_created_at (`created_at`);
ALTER TABLE `assets` ADD INDEX idx_project (`project`);
ALTER TABLE `vulnerabilities` ADD INDEX idx_cwe_id (`cwe_id`);
ALTER TABLE `vulnerabilities` ADD INDEX idx_severity (`severity`);
ALTER TABLE `vulnerability_findings` ADD INDEX idx_scan_id (`scan_id`);
ALTER TABLE `vulnerability_findings` ADD INDEX idx_vulnerability_id (`vulnerability_id`);

-- =====================================================
-- VULNERABILITY DATA (134+ vulnerability types)
-- =====================================================

INSERT INTO vulnerabilities(cwe_id, name, severity, score) VALUES
(0,'Cloud Metadata Potentially Exposed','high',1)
,(0,'User Agent Fuzzer','low',0)
,(0,'Script Active Scan Rules','low',0)
,(0,'SOAP Action Spoofing','high',1)
,(0,'SOAP XML Injection','high',1)
,(0,'Bypassing 403','low',0)
,(16,'GET for POST','medium',0)
,(20,'Source Code Disclosure - CVE-2012-1823','critical',4)
,(20,'Remote Code Execution - CVE-2012-1823','critical',20)
,(20,'Httpoxy - Proxy Header Misuse','medium',0)
,(20,'Relative Path Confusion','medium',0)
,(20,'HTTP Parameter Pollution','medium',0)
,(22,'Path Traversal','high',4)
,(74,'Server Side Template Injection (Blind)','high',4)
,(78,'Remote OS Command Injection','critical',20)
,(78,'Remote Code Execution - Shell Shock','critical',8)
,(78,'Spring4Shell','critical',8)
,(79,'Cross Site Scripting (Reflected)','medium',0)
,(79,'Cross Site Scripting (Persistent)','high',1)
,(79,'Cross Site Scripting (DOM Based)','medium',0)
,(79,'Cross Site Scripting (Persistent) - Prime','high',1)
,(79,'Cross Site Scripting (Persistent) - Spider','high',1)
,(79,'Out of Band XSS','low',0)
,(89,'SQL Injection','critical',20)
,(89,'SQL Injection - MySQL','critical',20)
,(89,'SQL Injection - Hypersonic SQL','critical',20)
,(89,'SQL Injection - Oracle','critical',20)
,(89,'SQL Injection - PostgreSQL','critical',20)
,(89,'SQL Injection - SQLite','critical',20)
,(89,'SQL Injection - MsSQL','critical',20)
,(91,'XSLT Injection','high',1)
,(94,'Server Side Code Injection','critical',1)
,(94,'ELMAH Information Leak','medium',0)
,(94,'.htaccess Information Leak','medium',0)
,(94,'Server Side Template Injection','critical',4)
,(97,'Server Side Include','high',2)
,(98,'Remote File Inclusion','high',4)
,(113,'CRLF Injection','high',1)
,(117,'Log4Shell','critical',8)
,(119,'Heartbleed OpenSSL Vulnerability','critical',1)
,(120,'Buffer Overflow','high',1)
,(134,'Format String Error','high',1)
,(190,'Integer Overflow Error','medium',0)
,(200,'Proxy Disclosure','low',0)
,(200,'Insecure HTTP Method','low',0)
,(200,'Possible Username Enumeration','medium',0)
,(200,'Cookie Slack Detector','medium',0)
,(209,'Generic Padding Oracle','high',1)
,(215,'Trace.axd Information Leak','medium',0)
,(215,'.env Information Leak','medium',0)
,(215,'Spring Actuator Information Leak','low',0)
,(264,'Cross-Domain Misconfiguration','medium',0)
,(311,'HTTP Only Site','low',0)
,(311,'HTTPS Content Available via HTTP','medium',0)
,(352,'Anti-CSRF Tokens Check','low',0)
,(384,'Session Fixation','high',1)
,(472,'Parameter Tampering','medium',0)
,(530,'Backup File Disclosure','high',1)
,(538,'Hidden File Finder','medium',0)
,(541,'Source Code Disclosure - /WEB-INF folder','high',1)
,(541,'Source Code Disclosure - Git','critical',20)
,(541,'Source Code Disclosure - File Inclusion','critical',8)
,(541,'Source Code Disclosure - SVN','critical',8)
,(548,'Directory Browsing','high',1)
,(601,'External Redirect','medium',0)
,(611,'XML External Entity Attack','high',1)
,(643,'XPath Injection','high',1)
,(776,'Exponential Entity Expansion (Billion Laughs Attack)','medium',0)
,(917,'Expression Language Injection','high',1)
,(942,'CORS Header','low',0);

-- =====================================================
-- BASE CONFIGURATION DATA
-- =====================================================

INSERT INTO configurations (`key`, `value`) VALUES
('scan_timeout_minutes', '10'),
('max_concurrent_scans', '5'),
('default_spider_depth', '2'),
('pass_threshold_score', '8'),
('api_version', '1.0.0');

-- =====================================================
-- VERIFICATION QUERIES (Check setup)
-- =====================================================

-- Verify users were created
SELECT CONCAT('✅ User created: ', User, '@', Host) as user_status 
FROM mysql.user WHERE User IN ('dast_user', 'dast_readonly');

-- Verify database exists
SHOW DATABASES LIKE 'dastapi';

-- Verify tables were created
SELECT CONCAT('✅ Tables created: ', COUNT(*), ' tables') as table_status 
FROM information_schema.tables WHERE table_schema = 'dastapi';

-- Verify vulnerabilities were loaded
SELECT CONCAT('✅ Vulnerabilities loaded: ', COUNT(*), ' vulnerability types') as vuln_status 
FROM vulnerabilities;

-- Verify configurations were loaded
SELECT CONCAT('✅ Configurations loaded: ', COUNT(*), ' settings') as config_status 
FROM configurations;