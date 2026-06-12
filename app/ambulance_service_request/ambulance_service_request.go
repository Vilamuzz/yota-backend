package ambulance_service_request

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/google/uuid"
)

type AmbulanceServiceRequest struct {
	ID               uuid.UUID                         `json:"id" gorm:"primaryKey"`
	AccountID        uuid.UUID                         `json:"accountId" gorm:"not null"`
	AmbulanceID      *uuid.UUID                        `json:"ambulanceId"`
	ApplicantName    string                            `json:"applicantName"`
	ApplicantPhone   string                            `json:"applicantPhone"`
	ApplicantAddress string                            `json:"applicantAddress"`
	Description      string                            `json:"description"`
	RequestDate      time.Time                         `json:"requestDate"`
	RequestReason    string                            `json:"requestReason"`
	Status           Status                            `json:"status"`
	ServiceCategory  ambulance_history.ServiceCategory `json:"serviceCategory"`
	RejectionReason  string                            `json:"rejectionReason"`
	CreatedAt        time.Time                         `json:"createdAt"`
	UpdatedAt        time.Time                         `json:"updatedAt"`

	Ambulance *ambulance.Ambulance `json:"ambulance" gorm:"foreignKey:AmbulanceID"`
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
	StatusInService Status = "in_service"
	StatusDone      Status = "done"
)
