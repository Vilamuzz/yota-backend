package social_program_subscription

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/google/uuid"
)

type SocialProgramSubscription struct {
	ID              uuid.UUID `json:"id" gorm:"primaryKey"`
	SocialProgramID uuid.UUID `json:"social_program_id" gorm:"not null"`
	AccountID       uuid.UUID `json:"account_id" gorm:"not null"`
	Status          Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Amount          float64   `json:"amount" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	SocialProgram *social_program.SocialProgram `gorm:"foreignKey:SocialProgramID;references:ID"`
	Account       *account.Account              `gorm:"foreignKey:AccountID;references:ID"`
}

type Status string

const (
	StatusActive  Status = "active"
	StatusPaused  Status = "paused"
	StatusStopped Status = "stopped"
)
