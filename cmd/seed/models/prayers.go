package models

import (
	"fmt"

	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedPrayers(db *gorm.DB) error {
	fmt.Println("Seeding prayers...")

	var programs []donation_program.DonationProgram
	if err := db.Find(&programs).Error; err != nil {
		return fmt.Errorf("failed to fetch donation programs: %w", err)
	}

	prayerContents := []string{
		"Semoga bantuan ini bermanfaat dan membawa berkah bagi sesama. Amin.",
		"Semoga berkah, berkah untuk donatur, berkah untuk penerima manfaat.",
		"Semoga lekas sembuh untuk adik-adik yang sedang berjuang melawan penyakit.",
		"Bismillah, semoga dilancarkan segala urusan bagi kita semua.",
		"Semoga program ini terus berjalan dan membantu lebih banyak orang.",
		"Doa terbaik untuk yayasan dan para donatur yang berhati mulia.",
		"Semoga menjadi amalan jariyah yang tidak terputus bagi donatur.",
		"Semoga Allah membalas kebaikan para donatur dengan berlipat ganda.",
		"Bismillah, dilancarkan rezekinya untuk semua yang telah berdonasi.",
		"Semoga Indonesia bebas stunting dan anak-anak tumbuh sehat.",
	}

	for _, program := range programs {
		var transactions []donation_program_transaction.DonationProgramTransaction
		err := db.Where("donation_program_id = ?", program.ID).Find(&transactions).Error
		if err != nil {
			return fmt.Errorf("failed to fetch transactions for program %s: %w", program.Title, err)
		}

		fmt.Printf("Seeding prayers for program: %s...\n", program.Title)

		for i, tx := range transactions {
			pID := uuid.New()
			isPublished := tx.TransactionStatus == "settlement"
			reportedVal := false

			pr := prayer.Prayer{
				ID:                           pID,
				DonationProgramTransactionID: tx.ID,
				Content:                      prayerContents[i%len(prayerContents)],
				IsPublished:                  isPublished,
				CreatedAt:                    tx.CreatedAt,
				Reported:                     &reportedVal,
				AmenCount:                    int64(i * 3),
				ReportCount:                  0,
			}

			// Check if already exists for this transaction
			var existing prayer.Prayer
			err := db.Where("donation_program_transaction_id = ?", pr.DonationProgramTransactionID).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&pr).Error; err != nil {
						return fmt.Errorf("failed to create prayer for transaction %s: %w", tx.OrderID, err)
					}
				} else {
					return fmt.Errorf("failed to check existing prayer: %w", err)
				}
			}
		}
	}

	return nil
}
