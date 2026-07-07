package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_transaction"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSocialProgramTransactions(db *gorm.DB) error {
	fmt.Println("Seeding social program transactions...")

	var paidInvoices []social_program_invoice.SocialProgramInvoice
	if err := db.Preload("Subscription").Where("status = ?", social_program_invoice.InvoiceStatusPaid).Find(&paidInvoices).Error; err != nil {
		return fmt.Errorf("failed to fetch paid invoices: %w", err)
	}

	for i, inv := range paidInvoices {
		if inv.Subscription == nil {
			continue
		}

		orderID := fmt.Sprintf("SPI-%s-%d", inv.ID.String()[:8], i+1)
		txID := uuid.New()
		paidAtTime := inv.DueDate.Add(-time.Hour * 12)

		isOnline := i%2 == 0
		var provider string
		var snapToken, snapRedirect string
		if isOnline {
			provider = "midtrans"
			snapToken = fmt.Sprintf("snap-token-spi-%s", uuid.New().String()[:8])
			snapRedirect = "https://app.sandbox.midtrans.com/snap/v2/vtweb/" + snapToken
		} else {
			provider = "offline"
		}

		var fee, netAmount float64
		var ppnPercentage, ppnAmount float64
		if isOnline {
			ppnPercentage = 11.0
			baseFee := 2000.0
			ppnAmount = baseFee * (ppnPercentage / 100.0)
			fee = baseFee
			netAmount = inv.MinimumAmount - (baseFee + ppnAmount)
			if netAmount < 0 {
				netAmount = 0
			}
		} else {
			fee = 0
			netAmount = inv.MinimumAmount
		}

		tx := social_program_transaction.SocialProgramTransaction{
			ID:                     txID,
			SocialProgramInvoiceID: inv.ID,
			OrderID:                orderID,
			AccountID:              inv.Subscription.AccountID,
			IsOnline:               isOnline,
			GrossAmount:            inv.MinimumAmount,
			FraudStatus:            "accept",
			TransactionStatus:      "settlement",
			Provider:               provider,
			TransactionID:          uuid.New().String(),
			SnapToken:              snapToken,
			SnapRedirectURL:        snapRedirect,
			Fee:                    fee,
			NetAmount:              netAmount,
			PpnPercentage:          ppnPercentage,
			PpnAmount:              ppnAmount,
			PaidAt:                 &paidAtTime,
			CreatedAt:              paidAtTime,
			UpdatedAt:              paidAtTime,
		}

		// Check if already exists
		var existing social_program_transaction.SocialProgramTransaction
		err := db.Where("social_program_invoice_id = ?", tx.SocialProgramInvoiceID).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tx).Error; err != nil {
					return fmt.Errorf("failed to create social program transaction for invoice %s: %w", inv.ID, err)
				}

				// Create corresponding FinanceRecord (FundID is the invoice ID as seen in the service)
				financeID := uuid.New().String()
				record := finance_record.FinanceRecord{
					ID:              financeID,
					FundType:        finance_record.FundTypeSocialProgram,
					FundID:          tx.SocialProgramInvoiceID.String(),
					SourceType:      finance_record.SourceTypeTransaction,
					SourceID:        tx.ID.String(),
					Amount:          tx.NetAmount,
					TransactionDate: *tx.PaidAt,
					CreatedAt:       tx.CreatedAt,
				}
				if err := db.Create(&record).Error; err != nil {
					return fmt.Errorf("failed to create finance record for social transaction: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check existing social transaction: %w", err)
			}
		}
	}

	return nil
}
