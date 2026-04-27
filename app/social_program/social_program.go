package social_program

import (
	"time"

	"github.com/google/uuid"
)

type SocialProgram struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey"`
	Slug          string    `json:"slug" gorm:"not null"`
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description" gorm:"not null"`
	CoverImage    string    `json:"coverImage" gorm:"not null"`
	Status        Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	MinimumAmount float64   `json:"minimumAmount" gorm:"not null"`
	BillingDay    int       `json:"billingDay" gorm:"not null"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DeletedAt     time.Time `json:"deletedAt" gorm:"index"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusStopped   Status = "stopped"
	StatusDraft     Status = "draft"
)
