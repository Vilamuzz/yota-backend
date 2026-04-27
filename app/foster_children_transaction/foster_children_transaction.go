package foster_children_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type FosterChildrenTransaction struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey"`
	FosterChildrenID  uuid.UUID  `json:"fosterChildrenId" gorm:"not null"`
	OrderID           string     `json:"orderId" gorm:"uniqueIndex"`
	AccountID         uuid.UUID  `json:"accountId"` // can be anonymous
	DonorName         string     `json:"donorName"`
	DonorEmail        string     `json:"donorEmail"`
	IsOnline          bool       `json:"isOnline"`
	GrossAmount       float64    `json:"grossAmount"`
	FraudStatus       string     `json:"fraudStatus"`
	TransactionStatus string     `json:"transactionStatus"`
	Provider          string     `json:"provider"` // midtrans
	TransactionID     string     `json:"transactionId"`
	SnapToken         string     `json:"snapToken"`
	SnapRedirectURL   string     `json:"snapRedirectUrl"`
	PaidAt            *time.Time `json:"paidAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`

	Account *account.Account `gorm:"foreignKey:AccountID;references:ID"`
}
