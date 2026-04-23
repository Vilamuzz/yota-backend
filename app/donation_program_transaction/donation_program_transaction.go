package donation_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type DonationProgramTransaction struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey"`
	DonationProgramID uuid.UUID  `json:"donation_program_id" gorm:"index;not null"`
	OrderID           string     `json:"order_id" gorm:"uniqueIndex"`
	AccountID         uuid.UUID  `json:"account_id" gorm:"index"`
	DonorName         string     `json:"donor_name"`
	DonorEmail        string     `json:"donor_email"`
	IsOnline          bool       `json:"is_online"`
	GrossAmount       float64    `json:"gross_amount"`
	FraudStatus       string     `json:"fraud_status"`
	TransactionStatus string     `json:"transaction_status" gorm:"index"`
	Provider          string     `json:"provider"`
	TransactionID     string     `json:"transaction_id"`
	SnapToken         string     `json:"snap_token"`
	SnapRedirectURL   string     `json:"snap_redirect_url"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	Account *account.Account `json:"-" gorm:"foreignKey:AccountID;references:ID"`
}

type TransactionStatus string

const (
	TransactionStatusSettlement TransactionStatus = "settlement"
)
