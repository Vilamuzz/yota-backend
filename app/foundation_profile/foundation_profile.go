package foundation_profile

import (
	"time"

	"github.com/google/uuid"
)

type FoundationProfile struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	Image     string    `json:"image"`
	Type      string    `json:"type"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
