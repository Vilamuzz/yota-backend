package social_program

import "time"

type SocialProgram struct {
	ID            string    `json:"id" gorm:"primary_key"`
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description" gorm:"not null"`
	ImageURL      string    `json:"image_url" gorm:"not null"`
	Status        Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	IsSubscribed  bool      `json:"is_subscribed" gorm:"-"`
	MinimumAmount float64   `json:"minimum_amount" gorm:"not null"`
	BillingDay    int       `json:"billing_day" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     time.Time `json:"deleted_at"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusStopped   Status = "stopped"
	StatusDraft     Status = "draft"
)
