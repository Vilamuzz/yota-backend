package social_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/google/uuid"
)

type SocialProgramTransaction struct {
	ID                     uuid.UUID  `json:"id" gorm:"primaryKey"`
	SocialProgramInvoiceID uuid.UUID  `json:"social_program_invoice_id" gorm:"unique;not null;index"`
	OrderID                string     `json:"order_id" gorm:"unique"`
	AccountID              uuid.UUID  `json:"account_id" gorm:"not null"`
	IsOnline               bool       `json:"is_online"`
	GrossAmount            float64    `json:"gross_amount"`
	FraudStatus            string     `json:"fraud_status"`
	TransactionStatus      string     `json:"transaction_status"`
	Provider               string     `json:"provider"`
	TransactionID          string     `json:"transaction_id" gorm:"unique"`
	SnapToken              string     `json:"snap_token"`
	SnapRedirectURL        string     `json:"snap_redirect_url"`
	PaidAt                 *time.Time `json:"paid_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	SocialProgramInvoice *social_program_invoice.SocialProgramInvoice `gorm:"foreignKey:SocialProgramInvoiceID;references:ID"`
	Account              *account.Account                             `gorm:"foreignKey:AccountID;references:ID"`
}
