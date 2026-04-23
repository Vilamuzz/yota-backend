package social_program_invoice

import (
	"time"

	"github.com/google/uuid"
)

type SocialProgramInvoice struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey"`
	SubscriptionID uuid.UUID `json:"subscription_id" gorm:"uniqueIndex:idx_subscription_billing;not null"`
	BillingPeriod  time.Time `json:"billing_period" gorm:"uniqueIndex:idx_subscription_billing;not null"`
	Amount         float64   `json:"amount" gorm:"not null"`
	Status         Status    `json:"status" gorm:"index:idx_status_due_date;type:varchar(20);not null;default:'active'"`
	DueDate        time.Time `json:"due_date" gorm:"index:idx_status_due_date;not null"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Status string

const (
	StatusActive Status = "active"
	StatusPaid   Status = "paid"
	StatusUnpaid Status = "unpaid"
)
