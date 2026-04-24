package models

import (
	"fmt"
	"log"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedMockUsers(db *gorm.DB) error {
	fmt.Println("Seeding users...")

	// Default password for all seeded users
	defaultPassword := "Password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	type seedUser struct {
		Username      string
		Email         string
		RoleID        int
		Status        bool
		EmailVerified bool
	}

	users := []seedUser{
		{"chairman", "chairman@yota.com", 2, true, true},
		{"social_manager", "social@yota.com", 3, true, true},
		{"finance", "finance@yota.com", 4, true, true},
		{"ambulance_manager", "ambulance@yota.com", 5, true, true},
		{"publication_manager", "publication@yota.com", 6, true, true},
		{"user1", "user1@yota.com", 1, true, true},
		{"user2", "user2@yota.com", 1, true, true},
		{"user3", "user3@yota.com", 1, true, false},
		{"banned_user", "banned@yota.com", 1, false, true},
	}

	for _, u := range users {
		var existingAccount account.Account
		if err := db.Where("email = ?", u.Email).First(&existingAccount).Error; err == nil {
			fmt.Printf("⚠ User %s already exists, skipping...\n", u.Username)
			continue
		}

		acc := account.Account{
			ID:            uuid.New(),
			Email:         u.Email,
			Password:      string(hashedPassword),
			IsBanned:      !u.Status,
			EmailVerified: u.EmailVerified,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			UserProfile: account.UserProfile{
				ID:        uuid.New(),
				Username:  u.Username,
				Address:   "Jl. Contoh No. 123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AccountRoles: []account.AccountRole{
				{
					RoleID:    u.RoleID,
					IsDefault: true,
					IsActive:  true,
				},
			},
		}

		if err := db.Create(&acc).Error; err != nil {
			log.Printf("Warning: Failed to create user %s: %v", u.Username, err)
			continue
		}
		fmt.Printf("✓ Created user: %-20s | Email: %-25s | Role: %-20d\n", u.Username, u.Email, u.RoleID)

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
		fmt.Printf("%-20s | %-25s | %-20d | %-15s | %s\n",
			u.Username, u.Email, u.RoleID, verified, status)
	}

	fmt.Println("================================================================================")
	fmt.Println("\nYou can now login with any of these accounts using:")
	fmt.Println("  Email: [email from above]")
	fmt.Println("  Password: Password123")
	fmt.Println()

	return nil
}
