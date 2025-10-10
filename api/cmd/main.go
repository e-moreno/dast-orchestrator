package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"src/pkg/zapScanner"

	"src/cmd/config"
	"src/cmd/controller"
)

func main() {
	log.Println("[MAIN] Starting DAST API...")
	cfg := config.New()
	log.Println("[MAIN] Configuration loaded successfully")

	log.Println("[MAIN] Connecting to database (RO)...")
	dbConnRO, err := configDataBase(cfg.DBRO)
	if err != nil {
		log.Printf("[Main] Error connecting to DB: %s", err)
		log.Printf("[Main] Continuing without database connection...")
		dbConnRO = nil
	}
	log.Println("[MAIN] Database RO connection attempt completed")

	log.Println("[MAIN] Connecting to database (RW)...")
	dbConnRW, err := configDataBase(cfg.DBRW)
	if err != nil {
		log.Printf("[Main] Error connecting to DB: %s", err)
		log.Printf("[Main] Continuing without database connection...")
		dbConnRW = nil
	}
	log.Println("[MAIN] Database RW connection attempt completed")

	log.Println("[MAIN] Connecting to ZAP scanner...")
	zap, err := zapScanner.NewWithAuth(cfg.ZapURL, cfg.ZapAPIKey)
	if err != nil {
		log.Printf("Couldn't connect to zap, error: %v", err)
	}
	log.Println("[MAIN] ZAP connection attempt completed")

	clr := controller.New(cfg, zap, dbConnRO, dbConnRW)

	log.Println("[MAIN] Creating URL mappings...")

	r := controller.CreateURLMappingsProd(clr, cfg)
	log.Println("[MAIN] URL mappings created successfully")

	log.Println("[MAIN] Running server...")
	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[MAIN] Server running successfully")
}

func configDataBase(dbCreds config.RDSCredentials) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", dbCreds.Username,
		dbCreds.Password, dbCreds.Host, dbCreds.Port, dbCreds.DBClusterIdentifier)
	log.Printf("[configDataBase] Trying to Connect to %s as %s ...",
		dbCreds.DBClusterIdentifier, dbCreds.Username)
	db, err := sql.Open(dbCreds.Engine, connectionString)
	if err != nil {
		return db, err
	}
	log.Printf("[configDataBase] DB connection OK")

	// Check that the database is available and accessible
	err = db.Ping()
	if err != nil {
		log.Printf("[configDataBase] Error Pinging DB: %s", err)
		return db, err
	}
	log.Printf("[configDataBase] DB ping check success!")

	return db, nil
}
