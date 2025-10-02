package zapScanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	inProgressStr             = "10"
	finishedStr               = "100"
	inProgress                = 10
	mockScanID                = "0"
	mockScanIDInProgress      = "1"
	mockScanIDFailed          = "2"
	mockScanIDResponseFailure = "-1"
)

type mockedSpiderClient struct {
	scanRes   map[string]scannerResponse
	statusRes map[string]scannerResponse
}

type scannerResponse struct {
	r   map[string]interface{}
	err error
}

func readFile(fileName string) []byte {
	currentDir, err := os.Getwd()
	if err != nil {
		return []byte("")
	}
	fullFilePath := currentDir + "/mocks/" + fileName
	fileBytes, err := ioutil.ReadFile(fullFilePath)
	if err != nil {
		return []byte("")
	}
	return fileBytes
}

func (mSC mockedSpiderClient) Scan(url string, maxchildren string, recurse string, contextname string,
	subtreeonly string,
) (map[string]interface{}, error) {
	response := mSC.scanRes[url]
	return response.r, response.err
}

func (mSC mockedSpiderClient) Status(scanID string) (map[string]interface{}, error) {
	response := mSC.statusRes[scanID]
	return response.r, response.err
}

func (mSC mockedSpiderClient) Stop(scanid string) (map[string]interface{}, error) {
	return nil, nil
}

func (mSC mockedSpiderClient) SetOptionMaxDuration(i int) (map[string]interface{}, error) {
	return nil, nil
}

type mockedActiveScanClient struct {
	scanRes   map[string]scannerResponse
	statusRes map[string]scannerResponse
}

func (mASC mockedActiveScanClient) Scan(url string, recurse string, inscopeonly string, scanpolicyname string,
	method string, postdata string, contextid string,
) (map[string]interface{}, error) {
	response := mASC.scanRes[url]
	return response.r, response.err
}

func (mASC mockedActiveScanClient) Status(scanID string) (map[string]interface{}, error) {
	response := mASC.statusRes[scanID]
	return response.r, response.err
}

// These functions are not used yet.
func (mASC mockedActiveScanClient) Stop(scanid string) (map[string]interface{}, error) {
	return nil, nil
}

func (mASC mockedActiveScanClient) SetOptionMaxScanDurationInMins(i int) (map[string]interface{}, error) {
	return nil, nil
}

func (mASC mockedActiveScanClient) AlertsIds(scanid string) (map[string]interface{}, error) {
	return nil, nil
}

type mockedCoreClient struct {
	b   []byte
	err error
}

func (mCC mockedCoreClient) Jsonreport() ([]byte, error) {
	return mCC.b, mCC.err
}

func (mCC mockedCoreClient) NewSession(name string, overwrite string) (map[string]interface{}, error) {
	return nil, mCC.err
}

func (mCC mockedCoreClient) LoadSession(name string) (map[string]interface{}, error) {
	return nil, mCC.err
}

func (mCC mockedCoreClient) SaveSession(name string, overwrite string) (map[string]interface{}, error) {
	return nil, mCC.err
}

type mockedAlertClient struct {
	r   map[string]interface{}
	err error
}

func (mAC mockedAlertClient) Alerts(baseurl string, start string, count string, riskid string) (map[string]interface{}, error) {
	return mAC.r, mAC.err
}

type zapClientMock struct {
	mSC  mockedSpiderClient
	mASC mockedActiveScanClient
	mCC  mockedCoreClient
	mAC  mockedAlertClient
}

func (z zapClientMock) Spider() SpiderClient {
	return z.mSC
}

func (z zapClientMock) Ascan() ActiveScanClient {
	return z.mASC
}

func (z zapClientMock) Core() CoreClient {
	return z.mCC
}

func (z zapClientMock) Alert() AlertClient {
	return z.mAC
}

var mockSC = mockedSpiderClient{
	map[string]scannerResponse{
		"https://www.google.com": {
			r:   map[string]interface{}{"scan": mockScanID},
			err: nil,
		},
		"https://this-should-fail.com": {
			r:   nil,
			err: errors.New("something went terribly wrong"),
		},
	},
	map[string]scannerResponse{
		"0": {
			r:   map[string]interface{}{"status": finishedStr},
			err: nil,
		},
		"1": {
			r:   map[string]interface{}{"status": inProgressStr},
			err: nil,
		},
		"2": {
			r:   nil,
			err: errors.New("something went terribly wrong"),
		},
	},
}

var mockASC = mockedActiveScanClient{
	scanRes: map[string]scannerResponse{
		"https://www.google.com": {
			r:   map[string]interface{}{"scan": mockScanID},
			err: nil,
		},
		"https://this-should-fail.com": {
			r:   nil,
			err: errors.New("something went terribly wrong"),
		},
	},
	statusRes: map[string]scannerResponse{
		mockScanID: {
			r:   map[string]interface{}{"status": finishedStr},
			err: nil,
		},
		mockScanIDInProgress: {
			r:   map[string]interface{}{"status": inProgressStr},
			err: nil,
		},
		mockScanIDFailed: {
			r:   nil,
			err: errors.New("something went terribly wrong"),
		},
	},
}

var mockAC = mockedAlertClient{
	r:   map[string]interface{}{},
	err: nil,
}

var mockCCSuccess = mockedCoreClient{
	readFile("scan_result.json"),
	nil,
}

var mockCCFailure = mockedCoreClient{
	nil,
	fmt.Errorf("error reading report results"),
}

var mockCCBadResults = mockedCoreClient{
	nil,
	nil,
}

func initMockService(mockCC mockedCoreClient) ZapService {
	var zapMock zapClient = zapClientMock{
		mSC:  mockSC,
		mASC: mockASC,
		mCC:  mockCC,
		mAC:  mockAC,
	}

	return ZapService{zapMock}
}

func TestStartScanOnSuccess(t *testing.T) {
	zapService := initMockService(mockCCSuccess)
	scanID, err := zapService.StartScan("https://www.google.com")

	assert.Equal(t, mockScanID, scanID)
	assert.Nil(t, err)
}

func TestStartScanOnFailure(t *testing.T) {
	zapService := initMockService(mockCCSuccess)
	scanID, err := zapService.StartScan("https://this-should-fail.com")

	assert.Equal(t, mockScanIDResponseFailure, scanID)
	assert.NotNil(t, err)
}

func TestCheckScanOnFinished(t *testing.T) {
	zapService := initMockService(mockCCSuccess)

	progress, _, err := zapService.CheckScan(mockScanID)

	assert.Nil(t, err)
	assert.Equal(t, finished, progress)
}

func TestCheckScanOnInProgressStatus(t *testing.T) {
	zapService := initMockService(mockCCFailure)

	progress, _, err := zapService.CheckScan(mockScanIDInProgress)

	assert.Nil(t, err)
	assert.Equal(t, inProgress, progress)
}

func TestCheckScanOnZapError(t *testing.T) {
	zapService := initMockService(mockCCFailure)

	_, _, err := zapService.CheckScan(mockScanIDFailed)

	assert.NotNil(t, err)
	assert.Equal(t,
		"error getting scan status: error getting active scan status: something went terribly wrong",
		err.Error())
}

func TestCheckScanOnCoreClientError(t *testing.T) {
	zapService := initMockService(mockCCFailure)
	_, err := zapService.StartScan("https://www.google.com")

	assert.Nil(t, err)

	_, _, err = zapService.CheckScan("0")

	assert.NotNil(t, err)
	assert.Equal(t, "error getting scan results: error reading report results", err.Error())
}

func TestCheckScanOnResultsError(t *testing.T) {
	zapService := initMockService(mockCCBadResults)
	_, err := zapService.StartScan("https://www.google.com")

	assert.Nil(t, err)

	_, _, err = zapService.CheckScan("0")

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
}
