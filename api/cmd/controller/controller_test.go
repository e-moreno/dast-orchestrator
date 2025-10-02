package controller

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"src/pkg/security"

	"src/pkg/zapScanner"

	"github.com/stretchr/testify/assert"

	"src/cmd/config"

	"github.com/DATA-DOG/go-sqlmock"
)

type zapSVMock struct {
	StartScanResponse string
	StartScanError    error

	CheckScanProgress int
	CheckScanResult   zapScanner.AScanResult
	CheckScanError    error
}

var cfg = config.New()

const mockHMACSecret = "5cdca760d1bccf301f765ed372028389652b70dd70256e69db28b2222792e21d"

func (z zapSVMock) StartScan(url string) (string, error) {
	return z.StartScanResponse, z.StartScanError
}

func (z zapSVMock) CheckScan(id string) (int, zapScanner.AScanResult, error) {
	return z.CheckScanProgress, z.CheckScanResult, z.CheckScanError
}

func (z zapSVMock) StartSession(id string) error {
	return nil
}

func (z zapSVMock) LoadSession(id string) error {
	return nil
}

func (z zapSVMock) SaveSession(id string) error {
	return nil
}

func (z zapSVMock) CheckScanAlerts(scanID string) (int, []zapScanner.FullAlert, error) {
	return 0, nil, nil
}

func (z zapSVMock) GetActiveScanAlerts(scanID string) (map[string]bool, error) {
	return nil, nil
}

func TestHealthCheckOk(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 0,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	db, mock, _ := sqlmock.New()
	sqlmock.MonitorPingsOption(true)
	mock.ExpectPing()
	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"api":"ok","dbro":"ok","dbrw":"ok","zap":"ok"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestHealthCheckOnZapFailure(t *testing.T) {
	zap := (*zapScanner.ZapService)(nil)
	db, mock, _ := sqlmock.New()
	sqlmock.MonitorPingsOption(true)
	mock.ExpectPing()
	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)
	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"api":"ok","dbro":"ok","dbrw":"ok","zap":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestCreateScanOnSuccess(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "1",
		StartScanError:    nil,
		CheckScanProgress: 0,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	db, mock, _ := sqlmock.New()
	cfg.HMACSecret = mockHMACSecret
	mock.ExpectExec("INSERT INTO scans(status, build_id, build_source, zap_id)" +
		" VALUES(\"started\", \"abcde-1234\", \"spinnaker\", \"1\") ")

	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)

	requestBody := ScanBody{
		BuildID:     "abcde-1234",
		Target:      "www.example.com",
		Application: "example",
		Source:      "spinnaker",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/scan", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := security.CalculateHMAC(requestBodyBytes, mockHMACSecret)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"scanID":"1","status":"started"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestCreateScanOnBadSignature(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "1",
		StartScanError:    nil,
		CheckScanProgress: 0,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	db, mock, _ := sqlmock.New()
	cfg.HMACSecret = "a1b2c3e4a1b2c3e4a1b2c3e4a1b2c3e4a1b2c3e4a1b2c3e4a1b2c3e4a1b2c3e4"
	mock.ExpectExec("INSERT INTO scans(status, build_id, build_source, zap_id)" +
		" VALUES(\"started\", \"abcde-1234\", \"spinnaker\", \"1\") ")

	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)

	requestBody := ScanBody{
		BuildID:     "abcde-1234",
		Target:      "www.example.com",
		Application: "example",
		Source:      "spinnaker",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/scan", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := security.CalculateHMAC(requestBodyBytes, mockHMACSecret)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusUnauthorized, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := ""
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestCreateScanOnZapError(t *testing.T) {
	zapErr := errors.New("failed starting scan")
	zap := zapSVMock{
		StartScanResponse: "1",
		StartScanError:    zapErr,
		CheckScanProgress: 0,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	cfg.HMACSecret = mockHMACSecret
	clr := New(cfg, zap, nil, nil)
	router := CreateURLMappings(clr, cfg)

	requestBody := ScanBody{
		BuildID:     "abcde-1234",
		Target:      "www.example.com",
		Application: "example",
		Source:      "spinnaker",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/scan", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := security.CalculateHMAC(requestBodyBytes, mockHMACSecret)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code, "Expected code %d, received code %d",
		http.StatusInternalServerError, response.Code)
	expectedResponse := `{"reason":"zap client error: failed starting scan","status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestCreateScanOnZapConnectionError(t *testing.T) {
	zap := (*zapScanner.ZapService)(nil)
	cfg.HMACSecret = mockHMACSecret
	clr := New(cfg, zap, nil, nil)
	router := CreateURLMappings(clr, cfg)

	requestBody := ScanBody{
		BuildID:     "abcde-1234",
		Target:      "www.example.com",
		Application: "example",
		Source:      "spinnaker",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/scan", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := security.CalculateHMAC(requestBodyBytes, mockHMACSecret)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code, "Expected code %d, received code %d",
		http.StatusInternalServerError, response.Code)
	expectedResponse := `{"reason":"not connected to zap scanner instance","status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestCreateScanOnBadRequest(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 0,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	cfg.HMACSecret = mockHMACSecret
	clr := New(cfg, zap, nil, nil)

	router := CreateURLMappings(clr, cfg)

	requestBodyBytes := []byte("{something invalid}")

	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/scan", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := security.CalculateHMAC(requestBodyBytes, mockHMACSecret)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code, "Expected code %d, received code %d",
		http.StatusBadRequest, response.Code)
	expectedResponse := `{"reason":"couldn't parse scan information from body","status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestGetScanInformationOnRunningScan(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 69,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    nil,
	}
	db, mock, _ := sqlmock.New()
	rows := sqlmock.NewRows([]string{
		"id", "scans.status", "scans.build_id", "scans.zap_id",
	}).AddRow("1", "running", "abcde-1234", "1")
	mock.ExpectQuery("SELECT id, status, build_id, zap_id " +
		"FROM scans WHERE build_id=?").WillReturnRows(rows)
	clr := New(cfg, zap, db, db)

	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/scan/abcde-1234", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"progress":69,"status":"running"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestGetScanInformationOnFinishedScanPassed(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 100,
		CheckScanResult: zapScanner.AScanResult{
			Version:   "",
			Generated: "",
			Sites: []zapScanner.Site{{
				Name: "",
				Host: "",
				Port: "",
				Ssl:  "",
				Alerts: []zapScanner.Alert{{
					Pluginid:   "",
					AlertRef:   "",
					Alert:      "",
					Name:       "Directory listing",
					Riskcode:   "",
					Confidence: "",
					Riskdesc:   "",
					Desc:       "",
					Instances: []zapScanner.Instance{
						{
							URI:      "/some-endpoint",
							Method:   "",
							Param:    "",
							Attack:   "",
							Evidence: "",
						},
					},
					Count:     "",
					Solution:  "",
					Otherinfo: "",
					Reference: "",
					Cweid:     "584",
					Wascid:    "",
					Sourceid:  "",
				}},
			}},
		},
		CheckScanError: nil,
	}
	db, mock, _ := sqlmock.New()
	rows := sqlmock.NewRows([]string{
		"id", "scans.status", "scans.build_id", "scans.zap_id",
	}).AddRow("2", "running", "abcde-1234", "1")
	mock.ExpectQuery("SELECT id, status, build_id, zap_id " +
		"FROM scans WHERE build_id=?").WillReturnRows(rows)

	clr := New(cfg, zap, db, db)

	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/scan/1", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"status":"passed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestGetScanInformationOnFinishedScanVulnerable(t *testing.T) {
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 100,
		CheckScanResult: zapScanner.AScanResult{
			Version:   "",
			Generated: "",
			Sites: []zapScanner.Site{{
				Name: "",
				Host: "",
				Port: "",
				Ssl:  "",
				Alerts: []zapScanner.Alert{
					{
						Pluginid:   "",
						AlertRef:   "",
						Alert:      "",
						Name:       "SQL Injection",
						Riskcode:   "",
						Confidence: "",
						Riskdesc:   "",
						Desc:       "",
						Instances: []zapScanner.Instance{
							{
								URI:      "/vulnerable-endpoint-1",
								Method:   "",
								Param:    "",
								Attack:   "",
								Evidence: "",
							},
							{
								URI:      "/vulnerable-endpoint-2",
								Method:   "",
								Param:    "",
								Attack:   "",
								Evidence: "",
							},
						},
						Count:     "",
						Solution:  "",
						Otherinfo: "",
						Reference: "",
						Cweid:     "89",
						Wascid:    "",
						Sourceid:  "",
					},
					{
						Pluginid:   "",
						AlertRef:   "",
						Alert:      "",
						Name:       ".git source code exposed",
						Riskcode:   "",
						Confidence: "",
						Riskdesc:   "",
						Desc:       "",
						Instances: []zapScanner.Instance{
							{
								URI:      "/vulnerable-endpoint-1",
								Method:   "",
								Param:    "",
								Attack:   "",
								Evidence: "",
							},
							{
								URI:      "/vulnerable-endpoint-2",
								Method:   "",
								Param:    "",
								Attack:   "",
								Evidence: "",
							},
						},
						Count:     "",
						Solution:  "",
						Otherinfo: "",
						Reference: "",
						Cweid:     "527",
						Wascid:    "",
						Sourceid:  "",
					},
				},
			}},
		},
		CheckScanError: nil,
	}
	db, mock, _ := sqlmock.New()
	rows := sqlmock.NewRows([]string{
		"id", "scans.status", "scans.build_id", "scans.zap_id",
	}).AddRow("3", "running", "abcde-1234", "1")
	mock.ExpectQuery("SELECT id, status, build_id, zap_id " +
		"FROM scans WHERE build_id=?").WillReturnRows(rows)

	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/scan/abcde-1234", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestGetScanInformationOnZapError(t *testing.T) {
	zapErr := errors.New("some zap error getting status")
	zap := zapSVMock{
		StartScanResponse: "",
		StartScanError:    nil,
		CheckScanProgress: 420,
		CheckScanResult:   zapScanner.AScanResult{},
		CheckScanError:    zapErr,
	}

	db, mock, _ := sqlmock.New()
	rows := sqlmock.NewRows([]string{
		"id", "scans.status", "scans.build_id", "scans.zap_id",
	}).AddRow("4", "running", "abcde-1234", "1")
	mock.ExpectQuery("SELECT id, status, build_id, zap_id " +
		"FROM scans WHERE build_id=?").WillReturnRows(rows)

	clr := New(cfg, zap, db, db)
	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/scan/abcde-1234", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
	expectedResponse := `{"reason":"error retrieving scan status: some zap error getting status","status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}

func TestGetScanInformationOnZapConnectionError(t *testing.T) {
	zap := (*zapScanner.ZapService)(nil)
	clr := New(cfg, zap, nil, nil)
	router := CreateURLMappings(clr, cfg)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/scan/1", nil)
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code, "Expected code %d, received code %d",
		http.StatusInternalServerError, response.Code)
	expectedResponse := `{"reason":"not connected to zap scanner instance","status":"failed"}`
	assert.Equal(t, expectedResponse, response.Body.String())
}
