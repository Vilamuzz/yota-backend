package auth

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID uuid.UUID `json:"account_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiredAt time.Time `json:"expired_at" gorm:"not null"`
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
}
