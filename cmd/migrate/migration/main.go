package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Vilamuzz/yota-backend/config"
	postgre_pkg "github.com/Vilamuzz/yota-backend/pkg/postgre"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	fmt.Println("Starting database migration...")

	// Connect to database
	db := config.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Get all models to migrate
	models := postgre_pkg.GetAllModels()

	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			fmt.Println("Running migrations...")
			if err := db.AutoMigrate(models...); err != nil {
				log.Fatalf("Failed to migrate tables: %v", err)
			}
			fmt.Println("Migration completed successfully!")
		case "drop":
			fmt.Println("Dropping all tables...")
			if err := db.Migrator().DropTable(models...); err != nil {
				log.Fatalf("Failed to drop tables: %v", err)
			}
			fmt.Println("All tables dropped successfully!")
		case "fresh":
			fmt.Println("Dropping all tables...")
			if err := db.Migrator().DropTable(models...); err != nil {
				log.Fatalf("Failed to drop tables: %v", err)
			}
			fmt.Println("Running fresh migration...")
			if err := db.AutoMigrate(models...); err != nil {
				log.Fatalf("Failed to migrate tables: %v", err)
			}
			fmt.Println("Fresh migration completed successfully!")
		default:
			fmt.Println("Unknown command. Available commands:")
			fmt.Println("  migrate - Run migrations")
			fmt.Println("  drop    - Drop all tables")
			fmt.Println("  fresh   - Drop all tables and re-run migrations")
			os.Exit(1)
		}
	} else {
		// Default behavior: just run migrations
		if err := db.AutoMigrate(models...); err != nil {
			log.Fatalf("Failed to migrate tables: %v", err)
		}
		fmt.Println("Migration completed successfully!")
	}
}
