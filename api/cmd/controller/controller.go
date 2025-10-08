package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"src/pkg/security"

	"src/pkg/scan"

	"src/cmd/config"
	"src/pkg/zapScanner"

	"github.com/gin-gonic/gin"
)

const (
	passed              = "passed"
	failed              = "failed"
	scanScoringTreshold = 20
)

type Controller struct {
	s     ScannerService
	c     *config.Configuration
	dbRO  *sql.DB
	dbRW  *sql.DB
	vulns map[string]scan.Vulnerability
}

type ScannerService interface {
	StartScan(string) (string, error)
	CheckScan(string) (int, zapScanner.AScanResult, error)
	CheckScanAlerts(scanID string) (int, []zapScanner.FullAlert, error)
	GetActiveScanAlerts(scanID string) (map[string]bool, error)
	StartSession(string) error
	LoadSession(string) error
	SaveSession(string) error
}

type ScanBody struct {
	BuildID     string `json:"build_id"`
	Target      string `json:"target"`
	Application string `json:"application"`
	Source      string `json:"source"`
}

type StatusBody struct {
	ScanID string
}

func New(cfg *config.Configuration, zapService ScannerService, dbRO *sql.DB, dbRW *sql.DB) *Controller {
	var v map[string]scan.Vulnerability
	var err error
	if dbRO != nil {
		v, err = scan.GetVulnerabilitiesFromDB(dbRO)
		if err != nil {
			log.Printf("Error reading vulnerabilities from DB: %v", err)
		}
	}
	c := Controller{zapService, cfg, dbRO, dbRW, v}
	return &c
}

func CreateURLMappings(clr *Controller, cfg *config.Configuration) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", clr.HealthCheck)

	r.POST("/scan", security.AuthMiddleware(cfg.HMACSecret), clr.CreateScanAsyncWait)

	r.POST("/status", security.AuthMiddleware(cfg.HMACSecret), clr.GetScanStatusFromDB)

	// Deprecated
	r.GET("/scan/:id", clr.GetScanStatus)

	r.POST("/reload", security.AuthMiddleware(cfg.HMACSecret), clr.Reload)

	return r
}

func CreateURLMappingsProd(clr *Controller, cfg *config.Configuration) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", clr.HealthCheck)

	r.POST("/scan", security.AuthMiddleware(cfg.HMACSecret), clr.CreateScanAsyncWait)

	r.POST("/status", security.AuthMiddleware(cfg.HMACSecret), clr.GetScanStatusFromDB)

	r.POST("/reload", security.AuthMiddleware(cfg.HMACSecret), clr.Reload)

	return r
}

func (cImpl *Controller) HealthCheck(c *gin.Context) {
	zapStatus, dbro, dbrw := "ok", "ok", "ok"
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		zapStatus = failed
	}
	if cImpl.dbRO != nil {
		err := cImpl.dbRO.Ping()
		if err != nil {
			dbro = failed
		}
	} else {
		dbro = failed
	}

	if cImpl.dbRW != nil {
		err := cImpl.dbRW.Ping()
		if err != nil {
			dbrw = failed
		}
	} else {
		dbrw = failed
	}

	c.JSON(http.StatusOK, gin.H{
		"api":  "ok",
		"zap":  zapStatus,
		"dbro": dbro,
		"dbrw": dbrw,
	})
}

func (cImpl *Controller) Reload(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"reason": "invalid JSON body",
		})
		return
	}

	// Reload secrets and env variables
	err := cImpl.c.LoadConfig()
	if err != nil {
		log.Printf("Error reloading config values: %v", err)
	}
	// If zap client is not connected
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		s, err := zapScanner.NewWithAuth(cImpl.c.ZapURL, cImpl.c.ZapAPIKey)
		if (s != (*zapScanner.ZapService)(nil)) && (err == nil) {
			cImpl.s = s
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "reloaded", "zap": cImpl.c.ZapURL})
}

func (cImpl *Controller) CreateScan(c *gin.Context) {
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		log.Printf("Trying to create scan without zap connection. Aborted.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "not connected to zap scanner instance",
		})
		return
	}
	var s ScanBody
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"reason": "couldn't parse scan information from body",
		})
		return
	}
	err := cImpl.s.StartSession(s.BuildID)
	if err != nil {
		log.Printf("Error loading session for scan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "zap client error: " + err.Error(),
		})
		return
	}
	l := log.Default()
	l.Printf("Received scan data %+v", s)
	scanID, err := cImpl.s.StartScan(s.Target)
	if err != nil {
		log.Printf("Error initiating scan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "zap client error: " + err.Error(),
		})
		return
	}
	var ss scan.Scan
	ss.Build_id = s.BuildID
	ss.Build_source = s.Source
	ss.Target = s.Target
	id, err := strconv.Atoi(scanID)
	if err != nil {
		log.Printf("Error parsing zap id: %v", err)
	}
	ss.Zap_id = id
	ss.Status = "started"
	ss.ID, err = scan.AddScanToDB(cImpl.dbRW, ss)
	if err != nil {
		log.Printf("Error adding scan to database: %v", err)
	}
	c.JSON(http.StatusOK, gin.H{"scanID": scanID, "status": "started"})
}

func (cImpl *Controller) CreateScanAsyncWait(c *gin.Context) {
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		log.Printf("Trying to create scan without zap connection. Aborted.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "not connected to zap scanner instance",
		})
		return
	}
	var s ScanBody
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"reason": "couldn't parse scan information from body",
		})
		return
	}
	/*err := cImpl.s.StartSession(s.BuildID)
	if err != nil {
		log.Printf("Error creating session for scan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "zap client error: " + err.Error(),
		})
		return
	}*/
	l := log.Default()
	l.Printf("Received scan data %+v", s)
	scanID, err := cImpl.s.StartScan(s.Target)
	if err != nil {
		log.Printf("Error initiating scan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"reason": "zap client error: " + err.Error(),
		})
		return
	}
	var ss scan.Scan
	ss.Build_id = s.BuildID
	ss.Build_source = s.Source
	ss.Target = s.Target
	id, err := strconv.Atoi(scanID)
	if err != nil {
		log.Printf("Error parsing zap id: %v", err)
	}
	ss.Zap_id = id
	ss.Status = "started"
	ss.ID, err = scan.AddScanToDB(cImpl.dbRW, ss)
	if err != nil {
		log.Printf("Error adding scan to database: %v", err)
	}
	go cImpl.waitForScan(cImpl.dbRW, ss)
	c.JSON(http.StatusOK, gin.H{"scanID": scanID, "status": "started"})
}

func (cImpl *Controller) GetScanStatus(c *gin.Context) {
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		log.Printf("Trying to get scan information without zap connection. Aborted.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed", "reason": "not connected to zap scanner instance",
		})
		return
	}
	scanID := ""
	if c.Request.Method == "POST" {
		var b StatusBody
		err := c.BindJSON(&b)
		if err != nil {
			log.Printf("Bad request body %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "failed", "reason": "Bad request body",
			})
			return
		}
		scanID = b.ScanID
	} else {
		scanID = c.Param("id")
	}

	s, err := scan.GetScanFromDB(cImpl.dbRO, scanID)
	if err != nil {
		log.Printf("Error readin scan from DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed", "reason": "zap client error: " + err.Error(),
		})
		return
	}
	progress, result, err := cImpl.s.CheckScan(strconv.Itoa(s.Zap_id))
	if err != nil {
		log.Printf("Error getting scan status: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"status": failed, "reason": "error retrieving scan status: " + err.Error(),
		})
		return
	}
	if progress < 100 {
		c.JSON(http.StatusOK, gin.H{
			"status":   "running",
			"progress": progress,
		})
		return
	}
	pass, err := scan.CheckScan(result)
	if err != nil {
		log.Printf("Error processing scan result: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"status": "failed", "reason": "error processing scan result: " + err.Error(),
		})
		return
	}
	if pass {
		c.JSON(http.StatusOK, gin.H{"status": passed})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": failed})
	}
}

func (cImpl *Controller) GetScanStatusFromDB(c *gin.Context) {
	if cImpl.s == (*zapScanner.ZapService)(nil) {
		log.Printf("Trying to get scan information without zap connection. Aborted.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed", "reason": "not connected to zap scanner instance",
		})
		return
	}
	var b StatusBody
	err := c.BindJSON(&b)
	if err != nil {
		log.Printf("Bad request body %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed", "reason": "Bad request body",
		})
		return
	}

	s, err := scan.GetScanFromDB(cImpl.dbRO, b.ScanID)
	if err != nil {
		log.Printf("Error readin scan from DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed", "reason": "zap client error: " + err.Error(),
		})
		return
	}
	if (s.Status == "failed") || (s.Status == "passed") || (s.Status == "error") || (s.Status == "started") {
		c.JSON(http.StatusOK, gin.H{
			"status": s.Status,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":   "running",
			"progress": s.Status,
		})
	}
}

func (cImpl *Controller) waitForScan(conn *sql.DB, s scan.Scan) {
	progress := 0
	scanID := strconv.Itoa(s.Zap_id)
	log.Printf("%+v\n", s)
	var result []zapScanner.FullAlert
	var err error
	for progress < 100 {
		progress, result, err = cImpl.s.CheckScanAlerts(scanID)
		if err != nil {
			log.Printf("Error getting scan status: %v", err)
		}
		err = scan.UpdateScanStatus(conn, strconv.Itoa(progress), s.Build_id)
		if err != nil {
			log.Printf("Error updating scan status: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	idsFromScan, err := cImpl.s.GetActiveScanAlerts(scanID)
	if err != nil {
		log.Printf("Failed to get active scan alert ids: %v", err)
	}
	r := checkAlerts(conn, result, s.ID, idsFromScan, cImpl.vulns)
	status := "failed"
	if r {
		status = "passed"
	}
	err = scan.UpdateScanStatus(conn, status, s.Build_id)
	if err != nil {
		log.Printf("Error updating scan status: %v", err)
	}
	/*err = cImpl.s.SaveSession(s.Build_id)
	if err != nil {
		log.Printf("Error saving session for scan: %s %v", s.Build_id, err)
	}*/
}

func checkAlerts(conn *sql.DB, alerts []zapScanner.FullAlert, scanID int64, ids map[string]bool,
	vulnerabilities map[string]scan.Vulnerability,
) bool {
	score := 0
	for _, a := range alerts {
		if _, ok := ids[a.ID]; ok {
			d := fmt.Sprintf("[Finding] CWE %s URL %s: %s \n", a.Cweid, a.URL, a.Description)
			v := vulnerabilities[a.Cweid]
			err := scan.AddVulnerabilityFinding(conn, scanID, v.ID, d)
			if err != nil {
				log.Printf("Error when saving findings to DB %v\n", err)
			}
			score += v.Score
		}
	}
	return score < 8
}
