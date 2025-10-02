package zapScanner

import (
	"github.com/zaproxy/zap-api-go/zap"
)

type ZapService struct {
	zapConn zapClient
}

type zapClient interface {
	Spider() SpiderClient
	Ascan() ActiveScanClient
	Core() CoreClient
	Alert() AlertClient
}

type zapClientImpl struct {
	zapConn zap.Interface
}

type SpiderClient interface {
	Scan(url string, maxchildren string, recurse string, contextname string,
		subtreeonly string) (map[string]interface{}, error)
	Status(scanid string) (map[string]interface{}, error)
	Stop(scanid string) (map[string]interface{}, error)
	SetOptionMaxDuration(i int) (map[string]interface{}, error)
}

type ActiveScanClient interface {
	Scan(url string, recurse string, inscopeonly string, scanpolicyname string, method string, postdata string,
		contextid string) (map[string]interface{}, error)
	Status(scanid string) (map[string]interface{}, error)
	Stop(scanid string) (map[string]interface{}, error)
	SetOptionMaxScanDurationInMins(i int) (map[string]interface{}, error)
	AlertsIds(scanid string) (map[string]interface{}, error)
}

type CoreClient interface {
	Jsonreport() ([]byte, error)
	NewSession(name string, overwrite string) (map[string]interface{}, error)
	LoadSession(name string) (map[string]interface{}, error)
	SaveSession(name string, overwrite string) (map[string]interface{}, error)
}

type AlertClient interface {
	Alerts(baseurl string, start string, count string, riskid string) (map[string]interface{}, error)
}

type FullAlert struct {
	Sourceid    string      `json:"sourceid"`
	Other       string      `json:"other"`
	Method      string      `json:"method"`
	Evidence    string      `json:"evidence"`
	PluginID    string      `json:"pluginId"`
	Cweid       string      `json:"cweid"`
	Confidence  string      `json:"confidence"`
	Wascid      string      `json:"wascid"`
	Description string      `json:"description"`
	MessageID   string      `json:"messageId"`
	InputVector string      `json:"inputVector"`
	URL         string      `json:"url"`
	Tags        interface{} `json:"tags"`
	Reference   string      `json:"reference"`
	Solution    string      `json:"solution"`
	Alert       string      `json:"alert"`
	Param       string      `json:"param"`
	Attack      string      `json:"attack"`
	Name        string      `json:"name"`
	Risk        string      `json:"risk"`
	ID          string      `json:"id"`
	AlertRef    string      `json:"alertRef"`
}

type AScanResult struct {
	Version   string `json:"@version"`
	Generated string `json:"@generated"`
	Sites     []Site `json:"site"`
}

type Site struct {
	Name   string  `json:"@name"`
	Host   string  `json:"@host"`
	Port   string  `json:"@port"`
	Ssl    string  `json:"@ssl"`
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Pluginid   string     `json:"pluginid"`
	AlertRef   string     `json:"alertRef"`
	Alert      string     `json:"alert"`
	Name       string     `json:"name"`
	Riskcode   string     `json:"riskcode"`
	Confidence string     `json:"confidence"`
	Riskdesc   string     `json:"riskdesc"`
	Desc       string     `json:"desc"`
	Instances  []Instance `json:"instances"`
	Count      string     `json:"count"`
	Solution   string     `json:"solution"`
	Otherinfo  string     `json:"otherinfo"`
	Reference  string     `json:"reference"`
	Cweid      string     `json:"cweid"`
	Wascid     string     `json:"wascid"`
	Sourceid   string     `json:"sourceid"`
}

type Instance struct {
	URI      string `json:"uri"`
	Method   string `json:"method"`
	Param    string `json:"param"`
	Attack   string `json:"attack"`
	Evidence string `json:"evidence"`
}

type ActiveScans struct {
	Scans []AScanData `json:"scans"`
}

type AScanData struct {
	ReqCount      string `json:"reqCount"`
	AlertCount    string `json:"alertCount"`
	Progress      string `json:"progress"`
	NewAlertCount string `json:"newAlertCount"`
	ID            string `json:"id"`
	State         string `json:"state"`
}
