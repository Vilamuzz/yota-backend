package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSocialProgramSubscriptions(db *gorm.DB) error {
	fmt.Println("Seeding social program subscriptions...")

	var programs []social_program.SocialProgram
	if err := db.Find(&programs).Error; err != nil {
		return fmt.Errorf("failed to fetch social programs: %w", err)
	}
	if len(programs) == 0 {
		return fmt.Errorf("no social programs found")
	}

	var users []account.Account
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("no users found")
	}

	now := time.Now()

	for i := 0; i < 10; i++ {
		subID := uuid.New()
		prog := programs[i%len(programs)]
		user := users[i%len(users)]

		sub := social_program_subscription.SocialProgramSubscription{
			ID:               subID,
			SocialProgramID:  prog.ID,
			AccountID:        user.ID,
			Status:           social_program_subscription.StatusActive,
			TotalPaidPeriods: 1, // We will make exactly 1 invoice paid, so 1 paid period
			CreatedAt:        now.AddDate(0, -10, 0), // started 10 months ago
			UpdatedAt:        now,
		}

		// Check if already exists for this program and user
		var existing social_program_subscription.SocialProgramSubscription
		err := db.Where("social_program_id = ? AND account_id = ?", sub.SocialProgramID, sub.AccountID).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&sub).Error; err != nil {
					return fmt.Errorf("failed to create subscription %d: %w", i+1, err)
				}
				fmt.Printf("✓ Created social program subscription: %s for %s\n", prog.Title, user.Email)
			} else {
				return fmt.Errorf("failed to check existing subscription: %w", err)
			}
		}
	}

	return nil
}