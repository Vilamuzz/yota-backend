package social_program

import (
	"time"

	"github.com/google/uuid"
)

type SocialProgram struct {
	ID               uuid.UUID        `json:"id" gorm:"primaryKey"`
	Slug             string           `json:"slug" gorm:"not null"`
	Title            string           `json:"title" gorm:"not null"`
	Description      string           `json:"description" gorm:"not null"`
	CoverImage       string           `json:"coverImage" gorm:"not null"`
	Status           Status           `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	SubmissionStatus SubmissionStatus `json:"submissionStatus" gorm:"type:varchar(20);not null;default:'diajukan'"`
	MinimumAmount    float64          `json:"minimumAmount" gorm:"not null"`
	BillingDay       int              `json:"billingDay" gorm:"not null"`
	RejectionReason  string           `json:"rejectionReason" gorm:"type:text"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	DeletedAt        *time.Time       `json:"deletedAt" gorm:"index"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusBerjalan Status = "berjalan"
	StatusSelesai  Status = "selesai"
)

type SubmissionStatus string

const (
	SubmissionDiajukan  SubmissionStatus = "diajukan"
	SubmissionDisetujui SubmissionStatus = "disetujui"
	SubmissionDitolak   SubmissionStatus = "ditolak"
)

type SocialProgramSubscription struct {
	ID              uuid.UUID          `json:"id" gorm:"primaryKey"`
	SocialProgramID uuid.UUID          `json:"socialProgramId" gorm:"not null"`
	AccountID       uuid.UUID          `json:"accountId" gorm:"not null"`
	Amount          float64            `json:"amount" gorm:"not null"`
	Status          SubscriptionStatus `json:"status" gorm:"type:varchar(20);not null;default:'belum_donasi'"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

type SubscriptionStatus string

const (
	SubscriptionSudahDonasi SubscriptionStatus = "sudah_donasi"
	SubscriptionBelumDonasi SubscriptionStatus = "belum_donasi"
	SubscriptionTidakAktif  SubscriptionStatus = "tidak_aktif"
)