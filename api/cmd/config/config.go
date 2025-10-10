package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

func getIntEnvOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("[Configuration] Error parsing %s as int: %v, using default %d", key, err, defaultValue)
		return defaultValue
	}
	return intValue
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
	zapHost := getEnvOrDefault("ZAP_HOST", "localhost")
	zapPort := getEnvOrDefault("ZAP_PORT", "8090")
	zapProto := getEnvOrDefault("ZAP_PROTO", "http")
	cfg.ZapURL = fmt.Sprintf("%s://%s:%s", zapProto, zapHost, zapPort)

	// Debug logging
	log.Printf("[LoadConfig] ZAP_HOST env var read as: '%s'", zapHost)
	log.Printf("[LoadConfig] ZAP_PORT env var read as: '%s'", zapPort)
	log.Printf("[LoadConfig] ZAP_PROTO env var read as: '%s'", zapProto)
	log.Printf("[LoadConfig] Constructed ZAP URL: '%s'", cfg.ZapURL)

	zapAPIKey := getEnvOrDefault("ZAP_KEY", "change-me-9203935709")
	cfg.ZapAPIKey = zapAPIKey

	HMACSecret := getEnvOrDefault("HMAC_SECRET", "736ffa5e4064da13711d075ed6b71069")
	cfg.HMACSecret = HMACSecret

	// Read shared database configuration
	dbEngine := getEnvOrDefault("DB_ENGINE", "mysql")
	dbHost := getEnvOrDefault("DB_HOST", "dast-db")
	dbPort := getIntEnvOrDefault("DB_PORT", 3306)
	dbCluster := getEnvOrDefault("DB_CLUSTER", "dastapi")

	// Read Database RO configuration
	var ro RDSCredentials
	ro.Username = getEnvOrDefault("DB_RO_USERNAME", "user")
	ro.Password = getEnvOrDefault("DB_RO_PASSWORD", "password")
	ro.Engine = dbEngine
	ro.Host = dbHost
	ro.Port = dbPort
	ro.DBClusterIdentifier = dbCluster

	// Read Database RW configuration  
	var rw RDSCredentials
	rw.Username = getEnvOrDefault("DB_RW_USERNAME", "user")
	rw.Password = getEnvOrDefault("DB_RW_PASSWORD", "password")
	rw.Engine = dbEngine
	rw.Host = dbHost
	rw.Port = dbPort
	rw.DBClusterIdentifier = dbCluster

	// Debug logging for database configuration
	log.Printf("[LoadConfig] DB RO: %s@%s:%d/%s", ro.Username, ro.Host, ro.Port, ro.DBClusterIdentifier)
	log.Printf("[LoadConfig] DB RW: %s@%s:%d/%s", rw.Username, rw.Host, rw.Port, rw.DBClusterIdentifier)

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
