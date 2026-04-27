package ambulance_request

import (
	"time"

	"github.com/google/uuid"
)

type AmbulanceRequest struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID        uuid.UUID `json:"accountId" gorm:"not null"`
	ApplicantName    string    `json:"applicantName" gorm:"not null"`
	ApplicantPhone   string    `json:"applicantPhone" gorm:"not null"`
	ApplicantAddress string    `json:"applicantAddress" gorm:"not null"`
	Description      string    `json:"description" gorm:"not null"`
	RequestDate      time.Time `json:"requestDate" gorm:"not null"`
	RequestReason    string    `json:"requestReason" gorm:"not null"`
	Status           Status    `json:"status" gorm:"not null"`
	RejectionReason  string    `json:"rejectionReason"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)
