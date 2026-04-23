package main

import (
	"flag"
	"fmt"
	"log"

	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/cmd/seed/models"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/google/uuid"
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
		if err := models.SeedMockUsers(db); err != nil {
			log.Fatalf("Failed to seed mock users: %v", err)
		}
	}

	fmt.Println("Seeding completed successfully!")
}

func seedSuperAdmin(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := account.Account{
		ID:            uuid.New(),
		Email:         "superadmin@yota.com",
		Password:      string(hashedPassword),
		IsBanned:      false,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UserProfile: account.UserProfile{
			ID:        uuid.New(),
			Username:  "superadmin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AccountRoles: []account.AccountRole{
			{
				RoleID:    8, // Superadmin role based on seedRoles
				IsDefault: true,
				IsActive:  true,
			},
			{
				RoleID:    1,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    2,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    3,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    4,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    5,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    6,
				IsDefault: false,
				IsActive:  true,
			},
			{
				RoleID:    7,
				IsDefault: false,
				IsActive:  true,
			},
		},
	}

	// Make sure we only create if user doesn't exist
	var existing account.Account
	// Just check by email to simplify
	if err := db.Where("email = ?", admin.Email).First(&existing).Error; err != nil {
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
