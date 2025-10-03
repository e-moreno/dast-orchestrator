package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Configuration struct {
	DBRO       RDSCredentials
	DBRW       RDSCredentials
	ZapAPIKey  string
	ZapURL     string
	HMACSecret string
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func New() *Configuration {
	var c Configuration
	err := c.LoadConfig()
	if err != nil {
		log.Printf("[Configuration][New] There was an error while reading the env variables: %v", err)
	}
	return &c
}

func (cfg *Configuration) LoadConfig() error {
	zapHost := getEnvOrDefault("ZAP_HOST", "zap-scanner")
	zapPort := getEnvOrDefault("ZAP_PORT", "8090")
	zapProto := getEnvOrDefault("ZAP_PROTO", "http")
	cfg.ZapURL = fmt.Sprintf("%s://%s:%s", zapProto, zapHost, zapPort)

	zapAPIKey := getEnvOrDefault("ZAP_KEY", "change-me-9203935709")
	cfg.ZapAPIKey = zapAPIKey

	HMACSecret := getEnvOrDefault("HMAC_SECRET", "736ffa5e4064da13711d075ed6b71069")
	cfg.HMACSecret = HMACSecret

	dbROEnv := getEnvOrDefault("DB_RO", "{\"username\":\"user\",\"password\":\"password\""+
		",\"engine\":\"mysql\",\"host\":\"dast-db\""+
		",\"port\":3306,\"dbClusterIdentifier\":\"dastapi\"}")
	dbRWEnv := getEnvOrDefault("DB_RW", "{\"username\":\"user\",\"password\":\"password\""+
		",\"engine\":\"mysql\",\"host\":\"dast-db\""+
		",\"port\":3306,\"dbClusterIdentifier\":\"dastapi\"}")

	var ro RDSCredentials
	var rw RDSCredentials

	err := json.Unmarshal([]byte(dbROEnv), &ro)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(dbRWEnv), &rw)
	if err != nil {
		return err
	}
	cfg.DBRO = ro
	cfg.DBRW = rw
	return nil
}

type RDSCredentials struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	DBClusterIdentifier string `json:"dbClusterIdentifier"`
}
