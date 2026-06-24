package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/app/foster_children_transaction"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedFosterChildrenTransactions(db *gorm.DB) error {
	fmt.Println("Seeding foster children transactions...")

	var children []foster_children.FosterChildren
	if err := db.Find(&children).Error; err != nil {
		return fmt.Errorf("failed to fetch foster children: %w", err)
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
		50000, 100000, 150000, 200000, 250000, 300000, 500000, 750000, 1000000, 1500000,
	}

	now := time.Now()

	for _, child := range children {
		fmt.Printf("Seeding transactions for foster child: %s...\n", child.Name)
		for i := 0; i < 10; i++ {
			txID := uuid.New()
			orderID := fmt.Sprintf("FOS-%s-%d-%s", child.ID.String()[:4], i+1, uuid.New().String()[:4])

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

			createdAt := now.AddDate(0, 0, -i)
			updatedAt := createdAt

			var paidAt *time.Time
			if status == "settlement" {
				paidTime := createdAt.Add(time.Minute * 15)
				paidAt = &paidTime
			}

			tx := foster_children_transaction.FosterChildrenTransaction{
				ID:                txID,
				FosterChildrenID:  child.ID,
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
			var existing foster_children_transaction.FosterChildrenTransaction
			err := db.Where("order_id = ?", tx.OrderID).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&tx).Error; err != nil {
						return fmt.Errorf("failed to create foster children transaction %s: %w", tx.OrderID, err)
					}

					// If transaction is settled, also create a FinanceRecord
					if status == "settlement" {
						financeID := uuid.New().String()
						record := finance_record.FinanceRecord{
							ID:              financeID,
							FundType:        finance_record.FundTypeFosterChildren,
							FundID:          tx.FosterChildrenID.String(),
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