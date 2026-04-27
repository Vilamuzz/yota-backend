package social_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/google/uuid"
)

type SocialProgramTransaction struct {
	ID                     uuid.UUID  `json:"id" gorm:"primaryKey"`
	SocialProgramInvoiceID uuid.UUID  `json:"socialProgramInvoiceId" gorm:"unique;not null;index"`
	OrderID                string     `json:"orderId" gorm:"unique"`
	AccountID              uuid.UUID  `json:"accountId" gorm:"not null"`
	IsOnline               bool       `json:"isOnline"`
	GrossAmount            float64    `json:"grossAmount"`
	FraudStatus            string     `json:"fraudStatus"`
	TransactionStatus      string     `json:"transactionStatus"`
	Provider               string     `json:"provider"`
	TransactionID          string     `json:"transactionId" gorm:"unique"`
	SnapToken              string     `json:"snapToken"`
	SnapRedirectURL        string     `json:"snapRedirectUrl"`
	PaidAt                 *time.Time `json:"paidAt"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`

	SocialProgramInvoice *social_program_invoice.SocialProgramInvoice `gorm:"foreignKey:SocialProgramInvoiceID;references:ID"`
	Account              *account.Account                             `gorm:"foreignKey:AccountID;references:ID"`
}
