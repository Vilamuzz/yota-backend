package prayer

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/user"
)

type Prayer struct {
	ID          string    `json:"id" gorm:"primary_key"`
	DonationID  string    `json:"donation_id" gorm:"not null"`
	UserID      *string   `json:"user_id"`
	Content     string    `json:"content" gorm:"not null"`
	AmenCount   int       `json:"amen_count" gorm:"default:0"`
	IsAmen      bool      `json:"is_amen" gorm:"-"`
	ReportCount int       `json:"report_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`

	User *user.User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type PrayerAmen struct {
	ID       string `json:"id" gorm:"primary_key"`
	PrayerID string `json:"prayer_id" gorm:"not null"`
	UserID   string `json:"user_id" gorm:"not null"`
}

type PrayerReport struct {
	ID       string `json:"id" gorm:"primary_key"`
	PrayerID string `json:"prayer_id" gorm:"not null"`
	UserID   string `json:"user_id" gorm:"not null"`
	Reason   string `json:"reason" gorm:"not null"`
}
