package donation_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/google/uuid"
)

type DonationProgramTransaction struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey"`
	DonationProgramID uuid.UUID  `json:"donationProgramId" gorm:"index:idx_transaction_composite,priority:1;not null"`
	OrderID           string     `json:"orderId" gorm:"uniqueIndex"`
	AccountID         *uuid.UUID `json:"accountId" gorm:"index:idx_transaction_composite,priority:2"`
	DonorName         string     `json:"donorName"`
	DonorEmail        string     `json:"donorEmail"`
	IsOnline          bool       `json:"isOnline"`
	GrossAmount       float64    `json:"grossAmount"`
	FraudStatus       string     `json:"fraudStatus"`
	TransactionStatus string     `json:"transactionStatus" gorm:"index:idx_transaction_composite,priority:3"`
	Provider          string     `json:"provider"`
	TransactionID     string     `json:"transactionId"`
	SnapToken         string     `json:"snapToken"`
	SnapRedirectURL   string     `json:"snapRedirectUrl"`
	Fee               float64    `json:"fee" gorm:"default:0"`
	NetAmount         float64    `json:"netAmount" gorm:"default:0"`
	PpnPercentage     float64    `json:"ppnPercentage" gorm:"type:numeric;default:0"`
	PpnAmount         float64    `json:"ppnAmount" gorm:"type:numeric;default:0"`
	PaidAt            *time.Time `json:"paidAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`

	Account         *account.Account                  `json:"-" gorm:"foreignKey:AccountID;references:ID"`
	DonationProgram *donation_program.DonationProgram `json:"-" gorm:"foreignKey:DonationProgramID;references:ID"`
}

type TransactionStatus string

const (
	TransactionStatusSettlement TransactionStatus = "settlement"
)
