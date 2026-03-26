package social_program_subscription

import "time"

type SocialProgramSubscription struct {
	ID              string    `json:"id" gorm:"primary_key"`
	SocialProgramID string    `json:"social_program_id" gorm:"not null"`
	UserID          string    `json:"user_id" gorm:"not null"`
	Status          Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at"`
}

type Status string

const (
	StatusActive  Status = "active"
	StatusPaused  Status = "paused"
	StatusStopped Status = "stopped"
)
