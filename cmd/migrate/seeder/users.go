package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedUsers(db *gorm.DB) error {
	fmt.Println("Seeding users...")

	// Default password for all seeded users
	defaultPassword := "Password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []user.User{
		// Superadmin
		{
			ID:            uuid.New(),
			Username:      "superadmin",
			Email:         "superadmin@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleSuperadmin,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Chairman
		{
			ID:            uuid.New(),
			Username:      "chairman",
			Email:         "chairman@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleChairman,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Social Manager
		{
			ID:            uuid.New(),
			Username:      "social_manager",
			Email:         "social@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleSocialManager,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Finance
		{
			ID:            uuid.New(),
			Username:      "finance",
			Email:         "finance@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleFinance,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Ambulance Manager
		{
			ID:            uuid.New(),
			Username:      "ambulance_manager",
			Email:         "ambulance@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleAmbulanceManager,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Publication Manager
		{
			ID:            uuid.New(),
			Username:      "publication_manager",
			Email:         "publication@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RolePublicationManager,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Regular User 1
		{
			ID:            uuid.New(),
			Username:      "user1",
			Email:         "user1@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleUser,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Regular User 2
		{
			ID:            uuid.New(),
			Username:      "user2",
			Email:         "user2@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleUser,
			Status:        true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Regular User 3 (unverified email)
		{
			ID:            uuid.New(),
			Username:      "user3",
			Email:         "user3@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleUser,
			Status:        true,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		// Banned User
		{
			ID:            uuid.New(),
			Username:      "banned_user",
			Email:         "banned@yota.com",
			Password:      string(hashedPassword),
			Role:          user.RoleUser,
			Status:        false,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, u := range users {
		// Check if user already exists
		var existingUser user.User
		if err := db.Where("email = ? OR username = ?", u.Email, u.Username).First(&existingUser).Error; err == nil {
			fmt.Printf("⚠ User %s already exists, skipping...\n", u.Username)
			continue
		}

		if err := db.Create(&u).Error; err != nil {
			log.Printf("Warning: Failed to create user %s: %v", u.Username, err)
			continue
		}
		fmt.Printf("✓ Created user: %-20s | Email: %-25s | Role: %-20s\n", u.Username, u.Email, u.Role)
	}

	fmt.Println("\n================================================================================")
	fmt.Println("                       SEEDED USER CREDENTIALS")
	fmt.Println("================================================================================")
	fmt.Println("Default Password for all users: Password123")
	fmt.Println("\nUser Accounts:")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, u := range users {
		verified := "✓ Verified"
		if !u.EmailVerified {
			verified = "✗ Not Verified"
		}
		status := "Active"
		if !u.Status {
			status = "Banned"
		}
		fmt.Printf("%-20s | %-25s | %-20s | %-15s | %s\n",
			u.Username, u.Email, u.Role, verified, status)
	}

	fmt.Println("================================================================================")
	fmt.Println("\nYou can now login with any of these accounts using:")
	fmt.Println("  Email: [email from above]")
	fmt.Println("  Password: Password123")
	fmt.Println()

	return nil
}
