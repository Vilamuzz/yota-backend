package transaction_donation

import (
	"time"
)

type TransactionDonation struct {
	ID              string     `json:"id" gorm:"primary_key"`
	DonationID      string     `json:"donation_id" gorm:"not null"`
	OrderID         string     `json:"order_id" gorm:"uniqueIndex"`
	UserID          string     `json:"user_id"` // can be anonymous
	DonorName       string     `json:"donor_name"`
	DonorEmail      string     `json:"donor_email"`
	Source          bool       `json:"source"` // true = online, false = offline
	GrossAmount     float64    `json:"gross_amount"`
	PaymentMethod   string     `json:"payment_method"`
	PaymentStatus   string     `json:"payment_status"`
	Provider        string     `json:"provider"` // midtrans
	TransactionID   string     `json:"transaction_id"`
	SnapToken       string     `json:"snap_token"`
	SnapRedirectURL string     `json:"snap_redirect_url"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
