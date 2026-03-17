package prayer

import "time"

type Prayer struct {
	ID          string    `json:"id" gorm:"primary_key"`
	DonationID  string    `json:"donation_id" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Content     string    `json:"content" gorm:"not null"`
	LikeCount   int       `json:"like_count" gorm:"default:0"`
	ReportCount int       `json:"report_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`
}
