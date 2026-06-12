package social_program_invoice

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/google/uuid"
)

type SocialProgramInvoice struct {
	ID             uuid.UUID     `json:"id" gorm:"primaryKey"`
	SubscriptionID uuid.UUID     `json:"subscriptionId" gorm:"uniqueIndex:idx_subscription_billing;not null"`
	BillingPeriod  time.Time     `json:"billingPeriod" gorm:"uniqueIndex:idx_subscription_billing;not null"`
	MinimumAmount  float64       `json:"amount" gorm:"not null"`
	Status         InvoiceStatus `json:"status" gorm:"index:idx_status_due_date;type:varchar(20);not null;default:'pending'"`
	DueDate        time.Time     `json:"dueDate" gorm:"index:idx_status_due_date;not null"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	SnapToken      string        `json:"snapToken" gorm:"->"`

	Subscription *social_program_subscription.SocialProgramSubscription `gorm:"foreignKey:SubscriptionID;references:ID"`
}

type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
)
