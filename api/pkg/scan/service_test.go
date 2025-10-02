package scan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"src/pkg/zapScanner"
)

func TestCheckScanPassed(t *testing.T) {
	r := zapScanner.AScanResult{
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
	}

	result, err := CheckScan(r)

	assert.Nil(t, err)
	assert.Equal(t, false, result)
}

func TestCheckCWENotFoundInDB(t *testing.T) {
	r := zapScanner.AScanResult{
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
					Name:       "non existent cwe",
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
					Cweid:     "1234",
					Wascid:    "",
					Sourceid:  "",
				},
				{
					Pluginid:   "",
					AlertRef:   "",
					Alert:      "",
					Name:       "non existent cwe",
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
					Cweid:     "69420",
					Wascid:    "",
					Sourceid:  "",
				},
			},
		}},
	}

	result, err := CheckScan(r)

	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestCheckScanVulnerable(t *testing.T) {
	r := zapScanner.AScanResult{
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
	}

	result, err := CheckScan(r)

	assert.Nil(t, err)
	assert.Equal(t, true, result)
}
