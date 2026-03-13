package main

import (
	"fmt"
	"log"

	"github.com/Vilamuzz/yota-backend/app/user"
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
	fmt.Println("Starting database seeder...")

	db := config.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	if err := SeedRoles(db); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}

	if err := SeedSuperAdmin(db); err != nil {
		log.Fatalf("Failed to seed super admin: %v", err)
	}

	fmt.Println("Seeding completed successfully!")
}

func SeedRoles(db *gorm.DB) error {
	roles := []user.Role{
		{ID: 1, Role: "user"},
		{ID: 2, Role: "admin"},
		{ID: 3, Role: "superadmin"},
	}

	for _, role := range roles {
		// Use FirstOrCreate to prevent duplicates if seeder is run multiple times
		if err := db.Where(user.Role{ID: role.ID}).Assign(role).FirstOrCreate(&role).Error; err != nil {
			return fmt.Errorf("failed to seed role %s: %w", role.Role, err)
		}
	}
	fmt.Println("✅ Roles seeded")
	return nil
}

func SeedSuperAdmin(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := user.User{
		Username:      "superadmin",
		Email:         "superadmin@example.com",
		Password:      string(hashedPassword),
		RoleID:        3, // superadmin role
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

	fmt.Println("⚠️ SuperAdmin already exists, skipping")
	return nil
}
