package foster_children_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type FosterChildrenTransaction struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey"`
	FosterChildrenID  uuid.UUID  `json:"foster_children_id" gorm:"not null"`
	OrderID           string     `json:"order_id" gorm:"uniqueIndex"`
	AccountID         uuid.UUID  `json:"account_id"` // can be anonymous
	DonorName         string     `json:"donor_name"`
	DonorEmail        string     `json:"donor_email"`
	IsOnline          bool       `json:"is_online"`
	GrossAmount       float64    `json:"gross_amount"`
	FraudStatus       string     `json:"fraud_status"`
	TransactionStatus string     `json:"transaction_status"`
	Provider          string     `json:"provider"` // midtrans
	TransactionID     string     `json:"transaction_id"`
	SnapToken         string     `json:"snap_token"`
	SnapRedirectURL   string     `json:"snap_redirect_url"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	Account *account.Account `gorm:"foreignKey:AccountID;references:ID"`
}
