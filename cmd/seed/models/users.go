package models

import (
	"fmt"
	"log"
	"time"

	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userSeed holds a user definition together with the role to assign after creation.
type userSeed struct {
	user   user.User
	roleID int8
}

func SeedMockUsers(db *gorm.DB) error {
	fmt.Println("Seeding users...")

	// Default password for all seeded users
	defaultPassword := "Password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	seeds := []userSeed{
		// Chairman
		{user: user.User{ID: uuid.New(), Username: "chairman", Email: "chairman@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 2},
		// Social Manager
		{user: user.User{ID: uuid.New(), Username: "social_manager", Email: "social@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 3},
		// Finance
		{user: user.User{ID: uuid.New(), Username: "finance", Email: "finance@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 4},
		// Ambulance Manager
		{user: user.User{ID: uuid.New(), Username: "ambulance_manager", Email: "ambulance@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 5, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 5},
		// Publication Manager
		{user: user.User{ID: uuid.New(), Username: "publication_manager", Email: "publication@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 7, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 7},
		// Regular User 1
		{user: user.User{ID: uuid.New(), Username: "user1", Email: "user1@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 1},
		// Regular User 2
		{user: user.User{ID: uuid.New(), Username: "user2", Email: "user2@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: true, DefaultRoleID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 1},
		// Regular User 3 (unverified email)
		{user: user.User{ID: uuid.New(), Username: "user3", Email: "user3@yota.com", Password: string(hashedPassword), Status: true, EmailVerified: false, DefaultRoleID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 1},
		// Banned User
		{user: user.User{ID: uuid.New(), Username: "banned_user", Email: "banned@yota.com", Password: string(hashedPassword), Status: false, EmailVerified: true, DefaultRoleID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, roleID: 1},
	}

	for _, s := range seeds {
		u := s.user

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

		// Assign the configured role in the join table
		userRole := user.UserRole{UserID: u.ID, RoleID: s.roleID}
		if err := db.Create(&userRole).Error; err != nil {
			log.Printf("Warning: Failed to assign role for user %s: %v", u.Username, err)
		}
	}

	fmt.Println("\n================================================================================")
	fmt.Println("                       SEEDED USER CREDENTIALS")
	fmt.Println("================================================================================")
	fmt.Println("Default Password for all users: Password123")
	fmt.Println("\nUser Accounts:")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, s := range seeds {
		verified := "✓ Verified"
		if !s.user.EmailVerified {
			verified = "✗ Not Verified"
		}
		status := "Active"
		if !s.user.Status {
			status = "Banned"
		}
		fmt.Printf("%-20s | %-25s | %-15s | %s\n",
			s.user.Username, s.user.Email, verified, status)
	}

	fmt.Println("================================================================================")
	fmt.Println("\nYou can now login with any of these accounts using:")
	fmt.Println("  Email: [email from above]")
	fmt.Println("  Password: Password123")
	fmt.Println()

	return nil
}
