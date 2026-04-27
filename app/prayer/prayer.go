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
	ID                           uuid.UUID `json:"id" gorm:"primaryKey"`
	DonationProgramTransactionID uuid.UUID `json:"donationProgramTransactionId" gorm:"uniqueIndex;not null"`
	Content                      string    `json:"content" gorm:"not null"`
	IsPublished                  bool      `json:"isPublished" gorm:"default:false;index"`
	CreatedAt                    time.Time `json:"createdAt" gorm:"not null"`
	DeletedAt                    time.Time `json:"deletedAt" gorm:"index;not null"`

	DonationProgramTransaction DonationProgramTransaction `json:"-" gorm:"foreignKey:DonationProgramTransactionID"`
	PrayerAmens                []PrayerAmen               `json:"amens" gorm:"foreignKey:PrayerID;references:ID"`
	PrayerReports              []PrayerReport             `json:"reports" gorm:"foreignKey:PrayerID;references:ID"`
}

type PrayerAmen struct {
	PrayerID  uuid.UUID `json:"prayerId" gorm:"primaryKey;not null"`
	AccountID uuid.UUID `json:"accountId" gorm:"primaryKey;not null"`
}

type PrayerReport struct {
	PrayerID  uuid.UUID `json:"prayerId" gorm:"primaryKey;not null"`
	AccountID uuid.UUID `json:"accountId" gorm:"primaryKey;not null"`
	Reason    string    `json:"reason" gorm:"not null"`
}
