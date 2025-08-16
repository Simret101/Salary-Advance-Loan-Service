package database

import (
	"SalaryAdvance/pkg/config"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	cfg := config.LoadConfig()

	
	dsn := cfg.DBURL
	if dsn == "" {
		dsn = cfg.DBDsn
	}
	if dsn == "" {
		log.Fatal("Database connection string (DB_URL or DB_DSN) is not set")
	}

	
	if !strings.Contains(dsn, "sslmode=") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
	}

	log.Printf("Connecting to database with DSN: %s", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\nDSN: %s", err, dsn)
	}

	log.Println("Database connected successfully")
	return db
}
