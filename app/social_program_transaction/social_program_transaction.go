package social_program_transaction

import "time"

type SocialProgramTransaction struct {
	ID                     string     `json:"id" gorm:"primary_key"`
	SocialProgramInvoiceID string     `json:"social_program_invoice_id" gorm:"not null"`
	OrderID                string     `json:"order_id" gorm:"uniqueIndex"`
	UserID                 string     `json:"user_id" gorm:"not null"`
	Source                 bool       `json:"source"` // true = online, false = offline
	GrossAmount            float64    `json:"gross_amount"`
	FraudStatus            string     `json:"fraud_status"`
	TransactionStatus      string     `json:"transaction_status"`
	Provider               string     `json:"provider"` // midtrans
	TransactionID          string     `json:"transaction_id"`
	SnapToken              string     `json:"snap_token"`
	SnapRedirectURL        string     `json:"snap_redirect_url"`
	PaidAt                 *time.Time `json:"paid_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}
