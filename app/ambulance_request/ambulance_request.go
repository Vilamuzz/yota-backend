package ambulance_request

import (
	"time"

	"github.com/google/uuid"
)

type AmbulanceRequest struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID        uuid.UUID `json:"account_id" gorm:"not null"`
	ApplicantName    string    `json:"applicant_name" gorm:"not null"`
	ApplicantPhone   string    `json:"applicant_phone" gorm:"not null"`
	ApplicantAddress string    `json:"applicant_address" gorm:"not null"`
	Description      string    `json:"description" gorm:"not null"`
	RequestDate      time.Time `json:"request_date" gorm:"not null"`
	RequestReason    string    `json:"request_reason" gorm:"not null"`
	Status           Status    `json:"status" gorm:"not null"`
	RejectionReason  string    `json:"rejection_reason"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)
