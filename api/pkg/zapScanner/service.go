package zapScanner

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/zaproxy/zap-api-go/zap"
)

const (
	// These constants will be moved to the app config
	timeStepMS      = 500
	passiveScanWait = 1000

	finished = 100

	scanMaxDurationMinutes = 2
	spiderMaxDepth         = "2"

	failedStatus = "failed"
	passedStatus = "passed"
)

func New(url string) (*ZapService, error) {
	cfg := &zap.Config{
		Proxy: url,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		log.Printf("error with zap client creation: %v", err)
		return nil, err
	}
	coreClient := client.Core()
	_, err = coreClient.Version() // Test zap connection
	if err != nil {
		log.Printf("error with zap connection: %v", err)
		return nil, err
	}
	return &ZapService{&zapClientImpl{client}}, nil
}

func NewWithAuth(url string, apiKey string) (*ZapService, error) {
	cfg := &zap.Config{
		Proxy:  url,
		APIKey: apiKey,
	}

	client, err := zap.NewClient(cfg)
	if err != nil {
		log.Printf("error with zap client creation: %v", err)
		return nil, err
	}
	coreClient := client.Core()
	_, err = coreClient.Version() // Test zap connection
	if err != nil {
		log.Printf("error with zap connection: %v", err)
		return nil, err
	}
	return &ZapService{&zapClientImpl{client}}, nil
}

func (z *zapClientImpl) Spider() SpiderClient {
	return z.zapConn.Spider()
}

func (z *zapClientImpl) Ascan() ActiveScanClient {
	return z.zapConn.Ascan()
}

func (z *zapClientImpl) Core() CoreClient {
	return z.zapConn.Core()
}

func (z *zapClientImpl) Alert() AlertClient {
	return z.zapConn.Alert()
}

func (z *ZapService) StartSession(id string) error {
	_, err := z.zapConn.Core().NewSession(id, "yes")
	return err
}

func (z *ZapService) LoadSession(id string) error {
	_, err := z.zapConn.Core().LoadSession(id)
	return err
}

func (z *ZapService) SaveSession(id string) error {
	_, err := z.zapConn.Core().SaveSession(id, "yes")
	return err
}

func (z *ZapService) StartScan(target string) (string, error) {
	return startScan(z.zapConn.Spider(), z.zapConn.Ascan(), target)
}

func (z *ZapService) CheckScan(scanID string) (int, AScanResult, error) {
	progress, err := checkScan(z.zapConn.Ascan(), scanID)
	var r AScanResult
	if err != nil {
		return -1, r, fmt.Errorf("error getting scan status: %v", err)
	}
	var report []byte
	if progress >= finished {
		log.Printf("Active Scan with ID: %s completed", scanID)
		report, err = z.zapConn.Core().Jsonreport()
		if err != nil {
			return finished, r, fmt.Errorf("error getting scan results: %v", err)
		}
		err := json.Unmarshal(report, &r)
		if err != nil {
			return finished, r, err
		}
	}
	return progress, r, nil
}

func (z *ZapService) CheckScanAlerts(scanID string) (int, []FullAlert, error) {
	progress, err := checkScan(z.zapConn.Ascan(), scanID)
	var fullAlerts []FullAlert
	if err != nil {
		return -1, fullAlerts, fmt.Errorf("error getting scan status: %v", err)
	}
	if progress >= finished {
		log.Printf("Active Scan with ID: %s completed", scanID)
		res, err := z.zapConn.Alert().Alerts("", "", "", "")
		alerts := res["alerts"].([]interface{})
		a, err := json.Marshal(alerts)
		if err != nil {
			log.Printf("Error marshalling alerts to json - %v", err)
		}
		if err := json.Unmarshal(a, &fullAlerts); err != nil {
			log.Printf("Error unmarshalling alerts from json -  %v", err)
		}
		if err != nil {
			return finished, fullAlerts, err
		}
	}
	return progress, fullAlerts, nil
}

func (z *ZapService) GetActiveScanAlerts(scanID string) (map[string]bool, error) {
	r, err := z.zapConn.Ascan().AlertsIds(scanID)
	ids := make(map[string]bool)
	alertsIds := r["alertsIds"].([]interface{})
	for _, v := range alertsIds {
		ids[v.(string)] = true
	}
	return ids, err
}

func startScan(sc SpiderClient, asc ActiveScanClient, target string) (string, error) {
	// Start spider scan of the target
	fmt.Println("Spider : " + target)
	resp, err := sc.Scan(target, spiderMaxDepth, "", "", "")
	if err != nil {
		return "-1", fmt.Errorf("error creating spider scan: %v", err)
	}
	// The scan now returns a scan id to support concurrent scanning
	scanID := resp["scan"].(string)
	for {
		time.Sleep(timeStepMS * time.Millisecond)
		resp, err = sc.Status(scanID)
		if err != nil {
			break
		}
		progress, err := strconv.Atoi(resp["status"].(string))
		if err != nil {
			break
		}
		if progress >= finished {
			break
		}
	}
	if err != nil {
		return "-1", fmt.Errorf("error waiting for spider scan: %v", err)
	}
	fmt.Println("Spider complete")

	// Give the passive scanner a chance to complete
	time.Sleep(passiveScanWait * time.Millisecond)
	asc.SetOptionMaxScanDurationInMins(scanMaxDurationMinutes)
	fmt.Println("Active scan : " + target)
	resp, err = asc.Scan(target, "True", "False", "", "", "",
		"")
	if err != nil {
		return "-1", fmt.Errorf("error starting active scan: %v", err)
	}
	v, ok := resp["scan"]
	if !ok {
		return "-1", fmt.Errorf("error starting active scan: %v couldn't get scan ID from zap: %v", err, resp)
	}
	// The scan now returns a scan id to support concurrent scanning
	scanID = v.(string)

	return scanID, nil
}

func checkScan(asc ActiveScanClient, scanID string) (int, error) {
	resp, err := asc.Status(scanID)
	if err != nil {
		return -1, fmt.Errorf("error getting active scan status: %v", err)
	}
	progress, err := strconv.Atoi(resp["status"].(string))
	if err != nil {
		return -1, fmt.Errorf("error converting status to int: %v", err)
	}
	log.Printf("Active Scan progress : %d\n", progress)
	return progress, nil
}
