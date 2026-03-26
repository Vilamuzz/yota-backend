package social_program_invoice

import "time"

type SocialProgramInvoice struct {
	ID             string    `json:"id" gorm:"primary_key"`
	SubscriptionID string    `json:"subscription_id" gorm:"not null"`
	Year           int       `json:"year" gorm:"not null"`
	Month          int       `json:"month" gorm:"not null"`
	MinimumAmount  float64   `json:"minimum_amount" gorm:"not null"`
	Status         Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      time.Time `json:"deleted_at"`
}

type Status string

const (
	StatusActive  Status = "active"
	StatusPaid    Status = "paid"
	StatusUnpaid  Status = "unpaid"
	StatusPartial Status = "partial"
)
