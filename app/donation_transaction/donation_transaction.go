package donation_transaction

import (
	"time"
)

type DonationTransaction struct {
	ID                string     `json:"id" gorm:"primary_key"`
	DonationID        string     `json:"donation_id" gorm:"not null"`
	OrderID           string     `json:"order_id" gorm:"uniqueIndex"`
	UserID            string     `json:"user_id"` // can be anonymous
	DonorName         string     `json:"donor_name"`
	DonorEmail        string     `json:"donor_email"`
	Source            bool       `json:"source"` // true = online, false = offline
	GrossAmount       float64    `json:"gross_amount"`
	FraudStatus       string     `json:"fraud_status"`
	TransactionStatus string     `json:"transaction_status"`
	Provider          string     `json:"provider"` // midtrans
	TransactionID     string     `json:"transaction_id"`
	SnapToken         string     `json:"snap_token"`
	SnapRedirectURL   string     `json:"snap_redirect_url"`
	PrayerContent     string     `json:"prayer_content" gorm:"column:prayer_content"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type TransactionStatus string

const (
	TransactionStatusSettlement TransactionStatus = "settlement"
)
