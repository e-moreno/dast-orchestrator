package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expectedZapURL = "https://zap-test-host:42069"
)

func TestLoadConfig(t *testing.T) {
	envVars := make(map[string]string)
	DBRO := RDSCredentials{
		Username:            "root",
		Password:            "testpwd1",
		Engine:              "mysql",
		Host:                "test-db.glovotest.test",
		Port:                3306,
		DBClusterIdentifier: "somedb-id1",
	}
	DBRW := RDSCredentials{
		Username:            "root",
		Password:            "testpwd2",
		Engine:              "mariadb",
		Host:                "test-db2.glovotest.test",
		Port:                3307,
		DBClusterIdentifier: "somedb-id2",
	}

	envVars["ZAP_HOST"] = "zap-test-host"
	envVars["ZAP_PORT"] = "42069"
	DBROEnv, _ := json.Marshal(DBRO)
	DBRWEnv, _ := json.Marshal(DBRW)
	envVars["DB_RO"] = string(DBROEnv)
	envVars["DB_RW"] = string(DBRWEnv)

	for k, v := range envVars {
		err := os.Setenv(k, v)
		if err != nil {
			t.Errorf("Could not set env variable %s", k)
		}
	}

	cfg := Configuration{}
	cfg.LoadConfig()

	// ZAP
	assert.Equal(t, expectedZapURL, cfg.ZapURL)

	// DB Conn RO
	assert.Equal(t, DBRO, cfg.DBRO)

	// DB Conn RW
	assert.Equal(t, DBRW, cfg.DBRW)
}
