package auth

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID uuid.UUID `json:"accountId" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiredAt time.Time `json:"expiredAt" gorm:"not null"`
	IsUsed    bool      `json:"isUsed" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
}
