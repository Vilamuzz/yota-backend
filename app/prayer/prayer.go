package prayer

import (
	"time"

	"github.com/google/uuid"
)

type DonationProgramTransaction struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	DonorName string
}

type Prayer struct {
	ID                    uuid.UUID `json:"id" gorm:"primaryKey"`
	DonationTransactionID uuid.UUID `json:"donation_program_transaction_id" gorm:"uniqueIndex;not null"`
	Content               string    `json:"content" gorm:"not null"`
	IsPublished           bool      `json:"is_published" gorm:"default:false;index"`
	CreatedAt             time.Time `json:"created_at" gorm:"not null"`
	DeletedAt             time.Time `json:"deleted_at" gorm:"index;not null"`

	DonationProgramTransaction DonationProgramTransaction `json:"-" gorm:"foreignKey:DonationTransactionID"`
	PrayerAmens                []PrayerAmen               `json:"amens" gorm:"foreignKey:PrayerID;references:ID"`
	PrayerReports              []PrayerReport             `json:"reports" gorm:"foreignKey:PrayerID;references:ID"`
}

type PrayerAmen struct {
	PrayerID  uuid.UUID `json:"prayer_id" gorm:"primaryKey;not null"`
	AccountID uuid.UUID `json:"account_id" gorm:"primaryKey;not null"`
}

type PrayerReport struct {
	PrayerID  uuid.UUID `json:"prayer_id" gorm:"primaryKey;not null"`
	AccountID uuid.UUID `json:"account_id" gorm:"primaryKey;not null"`
	Reason    string    `json:"reason" gorm:"not null"`
}
