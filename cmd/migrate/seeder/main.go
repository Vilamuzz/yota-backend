package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	fmt.Println("Starting database seeding...")

	// Connect to database
	db := config.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "users":
			if err := seedUsers(db); err != nil {
				log.Fatalf("Failed to seed users: %v", err)
			}
		case "all":
			if err := seedAll(db); err != nil {
				log.Fatalf("Failed to seed database: %v", err)
			}
		case "reset":
			if err := resetAndSeed(db); err != nil {
				log.Fatalf("Failed to reset and seed: %v", err)
			}
		default:
			fmt.Println("Unknown command. Available commands:")
			fmt.Println("  users - Seed only users")
			fmt.Println("  all   - Seed all data")
			fmt.Println("  reset - Delete all users and reseed")
			os.Exit(1)
		}
	} else {
		// Default: seed all
		if err := seedAll(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	}

	fmt.Println("Database seeding completed successfully!")
}

func seedAll(db *gorm.DB) error {
	if err := seedUsers(db); err != nil {
		return err
	}
	return seedCategories(db)
}

func resetAndSeed(db *gorm.DB) error {
	fmt.Println("Deleting all existing users...")
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}
	fmt.Println("All users deleted.")
	return seedUsers(db)
}
