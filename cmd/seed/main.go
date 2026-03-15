package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/cmd/seed/models"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	// Load env file if exists (for local running)
	_ = godotenv.Load()
}

func main() {
	mockData := flag.Bool("mock-data", false, "Seed the database with mock categories and test users")
	flag.Parse()

	fmt.Println("Starting database seeder...")

	db := config.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	if err := models.SeedRoles(db); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}

	if err := seedSuperAdmin(db); err != nil {
		log.Fatalf("Failed to seed super admin: %v", err)
	}

	if *mockData {
		fmt.Println("Mock data flag provided. Seeding categories and test users...")
		if err := models.SeedCategories(db); err != nil {
			log.Fatalf("Failed to seed categories: %v", err)
		}
		if err := models.SeedMockUsers(db); err != nil {
			log.Fatalf("Failed to seed mock users: %v", err)
		}
	}

	fmt.Println("Seeding completed successfully!")
}

func seedSuperAdmin(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := user.User{
		Username:      "superadmin",
		Email:         "superadmin@yota.com",
		Password:      string(hashedPassword),
		RoleID:        8, // Superadmin role based on seedRoles
		Status:        true,
		EmailVerified: true,
	}

	// Make sure we only create if user doesn't exist
	var existing user.User
	if err := db.Where("email = ? OR username = ?", admin.Email, admin.Username).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to seed superadmin: %w", err)
			}
			fmt.Println("✅ SuperAdmin seeded")
			return nil
		}
		return err
	}

	fmt.Println("⚠ SuperAdmin already exists, skipping...")
	return nil
}
