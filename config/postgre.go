package config

import (
	"fmt"
	"log"
	"os"

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
	fmt.Println("Connected to the database successfully!")
	return db
}
