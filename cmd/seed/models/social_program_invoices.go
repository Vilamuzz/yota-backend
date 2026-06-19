package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSocialProgramInvoices(db *gorm.DB) error {
	fmt.Println("Seeding social program invoices...")

	var subscriptions []social_program_subscription.SocialProgramSubscription
	if err := db.Preload("SocialProgram").Find(&subscriptions).Error; err != nil {
		return fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	now := time.Now()

	for _, sub := range subscriptions {
		fmt.Printf("Seeding 10 invoices for subscription %s...\n", sub.ID)
		billingDay := 10
		if sub.SocialProgram != nil && sub.SocialProgram.BillingDay > 0 {
			billingDay = sub.SocialProgram.BillingDay
		}

		minAmount := 50000.0
		if sub.SocialProgram != nil && sub.SocialProgram.MinimumAmount > 0 {
			minAmount = sub.SocialProgram.MinimumAmount
		}

		for i := 0; i < 10; i++ {
			// Months ago: from 10 months ago to 1 month ago
			monthsAgo := 10 - i
			billingMonth := now.AddDate(0, -monthsAgo, 0)
			billingPeriod := time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			dueDate := time.Date(billingMonth.Year(), billingMonth.Month(), billingDay, 23, 59, 59, 0, time.UTC)

			status := social_program_invoice.InvoiceStatusPending
			// Make the first invoice (oldest one, i.e. i == 0) paid
			if i == 0 {
				status = social_program_invoice.InvoiceStatusPaid
			} else if dueDate.Before(now) {
				// Past invoices that are not paid can be overdue
				status = social_program_invoice.InvoiceStatusOverdue
			}

			inv := social_program_invoice.SocialProgramInvoice{
				ID:             uuid.New(),
				SubscriptionID: sub.ID,
				BillingPeriod:  billingPeriod,
				MinimumAmount:  minAmount,
				Status:         status,
				DueDate:        dueDate,
				CreatedAt:      billingPeriod,
				UpdatedAt:      billingPeriod,
			}

			// Check if already exists for this subscription and period
			var existing social_program_invoice.SocialProgramInvoice
			err := db.Where("subscription_id = ? AND billing_period = ?", inv.SubscriptionID, inv.BillingPeriod).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&inv).Error; err != nil {
						return fmt.Errorf("failed to create invoice for subscription %s: %w", sub.ID, err)
					}
				} else {
					return fmt.Errorf("failed to check existing invoice: %w", err)
				}
			}
		}
	}

	return nil
}