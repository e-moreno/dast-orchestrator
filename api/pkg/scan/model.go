package scan

import (
	"database/sql"
	"strconv"
	"time"
)

type Scan struct {
	ID           int64
	Status       string
	Build_id     string
	Build_source string
	Target       string
	Zap_id       int
	Start_date   time.Time
	End_date     time.Time
	Asset_id     int64
	Created_at   time.Time
}

type Vulnerability struct {
	Score int
	ID    int64
}

func AddScanToDB(conn *sql.DB, s Scan) (int64, error) {
	q := "INSERT INTO scans(status, build_id, build_source, target, zap_id) VALUES (?, ?, ?, ?, ?)"
	res, err := conn.Exec(q, s.Status, s.Build_id, s.Build_source, s.Target, s.Zap_id)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func GetScanFromDB(conn *sql.DB, buildID string) (Scan, error) {
	s := Scan{}
	q := "SELECT id, status, build_id, zap_id FROM scans WHERE build_id=?"

	r, err := conn.Query(q, buildID)
	if err != nil {
		return s, err
	}
	if r.Next() {
		err = r.Scan(&s.ID, &s.Status, &s.Build_id, &s.Zap_id)
	}
	return s, err
}

func UpdateScanStatus(conn *sql.DB, status string, build_id string) error {
	q := "UPDATE scans SET status=? WHERE build_id=?"
	_, err := conn.Exec(q, status, build_id)
	return err
}

func GetVulnerabilitiesFromDB(conn *sql.DB) (map[string]Vulnerability, error) {
	q := "SELECT cwe_id, id, score FROM vulnerabilities"
	rows, err := conn.Query(q)
	vulnerabilities := make(map[string]Vulnerability)
	if err != nil {
		return vulnerabilities, err
	}
	for rows.Next() {
		var cweID int
		var v Vulnerability
		if err := rows.Scan(&cweID, &v.ID, &v.Score); err != nil {
			return vulnerabilities, err
		} else {
			vulnerabilities[strconv.Itoa(cweID)] = v
		}
	}
	return vulnerabilities, err
}

func AddVulnerabilityFinding(conn *sql.DB, scanID int64, vulnerabilityID int64, description string) error {
	q := "INSERT INTO vulnerability_findings(scan_id, vulnerability_id, details) VALUES (?, ?, ?)"
	_, err := conn.Exec(q, scanID, vulnerabilityID, description)
	return err
}
