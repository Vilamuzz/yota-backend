package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	uri := os.Getenv("DB")
	if uri == "" {
		log.Fatal("Database URI is not set in environment variables")
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Configure the underlying sql.DB connection pool.
	// Without this, sql.DB defaults to MaxIdleConns=2, which causes constant
	// reconnections under load — each one paying the full SCRAM/pbkdf2 cost.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	maxOpenConns := 25
	if val := os.Getenv("DB_MAX_OPEN_CONNS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxOpenConns = parsed
		}
	}

	maxIdleConns := 10
	if val := os.Getenv("DB_MAX_IDLE_CONNS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxIdleConns = parsed
		}
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)              // max simultaneous connections to Postgres
	sqlDB.SetMaxIdleConns(maxIdleConns)             // keep this many warm; SCRAM is paid once per idle conn
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // recycle connections every 30 min (avoids stale conn issues)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // evict idle conns unused for 5 min

	fmt.Println("Connected to the database successfully!")
	return db
}
