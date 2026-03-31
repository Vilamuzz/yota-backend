package prayer

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/user"
)

type Prayer struct {
	ID                    string    `json:"id" gorm:"primary_key"`
	DonationID            string    `json:"donation_id" gorm:"not null"`
	DonationTransactionID string    `json:"donation_transaction_id" gorm:"not null"`
	UserID                *string   `json:"user_id"`
	Content               string    `json:"content" gorm:"not null"`
	Status                bool      `json:"status" gorm:"default:false"` // true = published, false = pending
	AmenCount             int       `json:"amen_count" gorm:"default:0"`
	IsAmen                bool      `json:"is_amen" gorm:"-"`
	ReportCount           int       `json:"report_count" gorm:"default:0"`
	CreatedAt             time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"not null"`

	User          *user.User     `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PrayerAmes    []PrayerAmen   `json:"prayer_ames" gorm:"foreignKey:PrayerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PrayerReports []PrayerReport `json:"prayer_reports" gorm:"foreignKey:PrayerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
