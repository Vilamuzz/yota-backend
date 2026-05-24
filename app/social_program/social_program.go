package social_program

import (
	"time"

	"github.com/google/uuid"
)

type SocialProgram struct {
	ID              uuid.UUID  `json:"id" gorm:"primaryKey"`
	Slug            string     `json:"slug" gorm:"not null"`
	Title           string     `json:"title" gorm:"not null"`
	Description     string     `json:"description" gorm:"not null"`
	CoverImage      string     `json:"coverImage" gorm:"not null"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	MinimumAmount   float64    `json:"minimumAmount" gorm:"not null"`
	BillingDay      int        `json:"billingDay" gorm:"not null"`
	RejectionReason string     `json:"rejectionReason"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`

	TotalSubscribers int64   `json:"totalSubscribers" gorm:"->"`
	IsSubscribed     bool    `json:"isSubscribed" gorm:"->"`
	SubscriptionID   string  `json:"subscriptionId" gorm:"->"`
	CollectedFund    float64 `json:"collectedFund" gorm:"->"`
	TotalExpense     float64 `json:"totalExpense" gorm:"->"`
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusRejected  Status = "rejected"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusCompleted, StatusRejected:
		return true
	}
	return false
}
