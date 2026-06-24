package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedDonationTransactions(db *gorm.DB) error {
	fmt.Println("Seeding donation transactions...")

	var programs []donation_program.DonationProgram
	if err := db.Find(&programs).Error; err != nil {
		return fmt.Errorf("failed to fetch donation programs: %w", err)
	}

	var users []account.Account
	if err := db.Preload("UserProfile").Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	donorNames := []string{
		"Hamba Allah", "Budi Sudarsono", "Siti Aminah", "Ahmad Fauzi",
		"Dewi Lestari", "Rian Hidayat", "Indah Permata", "Agus Setiawan",
		"Lani Cahyani", "Yusuf Mansur",
	}
	donorEmails := []string{
		"anonymous@example.com", "budi@gmail.com", "siti@yahoo.com", "ahmad@outlook.com",
		"dewi@gmail.com", "rian@gmail.com", "indah@yahoo.com", "agus@gmail.com",
		"lani@gmail.com", "yusuf@gmail.com",
	}
	amounts := []float64{
		20000, 50000, 100000, 150000, 250000, 300000, 500000, 750000, 1000000, 2500000,
	}

	now := time.Now()

	for _, program := range programs {
		fmt.Printf("Seeding transactions for program: %s...\n", program.Title)
		for i := 0; i < 10; i++ {
			txID := uuid.New()
			orderID := fmt.Sprintf("DON-%s-%d-%s", program.ID.String()[:4], i+1, uuid.New().String()[:4])

			// Assign some to registered users
			var accIDPtr *uuid.UUID
			donorName := donorNames[i%len(donorNames)]
			donorEmail := donorEmails[i%len(donorEmails)]

			// Assign every 3rd transaction to a registered user
			if i%3 == 0 && len(users) > 0 {
				userIndex := (i * 2) % len(users)
				u := users[userIndex]
				accIDPtr = &u.ID
				if u.UserProfile.Username != "" {
					donorName = u.UserProfile.Username
				} else {
					donorName = u.Email
				}
				donorEmail = u.Email
			}

			isOnline := i%2 == 0
			var provider string
			var snapToken, snapRedirect string
			if isOnline {
				provider = "midtrans"
				snapToken = fmt.Sprintf("snap-token-%s", uuid.New().String()[:8])
				snapRedirect = "https://app.sandbox.midtrans.com/snap/v2/vtweb/" + snapToken
			} else {
				provider = "offline"
			}

			// Vary the statuses: 8 settlement, 1 pending, 1 expire
			status := "settlement"
			if i == 8 {
				status = "pending"
			} else if i == 9 {
				status = "expire"
			}

			createdAt := now.AddDate(0, 0, -i) // spread over the last 10 days
			updatedAt := createdAt

			var paidAt *time.Time
			if status == "settlement" {
				paidTime := createdAt.Add(time.Minute * 15)
				paidAt = &paidTime
			}

			tx := donation_program_transaction.DonationProgramTransaction{
				ID:                txID,
				DonationProgramID: program.ID,
				OrderID:           orderID,
				AccountID:         accIDPtr,
				DonorName:         donorName,
				DonorEmail:        donorEmail,
				IsOnline:          isOnline,
				GrossAmount:       amounts[i%len(amounts)],
				FraudStatus:       "accept",
				TransactionStatus: status,
				Provider:          provider,
				SnapToken:         snapToken,
				SnapRedirectURL:   snapRedirect,
				PaidAt:            paidAt,
				CreatedAt:         createdAt,
				UpdatedAt:         updatedAt,
			}

			// Check if already exists by OrderID
			var existing donation_program_transaction.DonationProgramTransaction
			err := db.Where("order_id = ?", tx.OrderID).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&tx).Error; err != nil {
						return fmt.Errorf("failed to create transaction %s: %w", tx.OrderID, err)
					}

					// If transaction is settled, also create a FinanceRecord
					if status == "settlement" {
						financeID := uuid.New().String()
						record := finance_record.FinanceRecord{
							ID:              financeID,
							FundType:        finance_record.FundTypeDonation,
							FundID:          tx.DonationProgramID.String(),
							SourceType:      finance_record.SourceTypeTransaction,
							SourceID:        tx.ID.String(),
							Amount:          tx.GrossAmount,
							TransactionDate: *tx.PaidAt,
							CreatedAt:       tx.CreatedAt,
						}
						if err := db.Create(&record).Error; err != nil {
							return fmt.Errorf("failed to create finance record for transaction %s: %w", tx.OrderID, err)
						}
					}
				} else {
					return fmt.Errorf("failed to check existing transaction: %w", err)
				}
			}
		}
	}

	return nil
}