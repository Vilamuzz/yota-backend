package ambulance_service_request

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/google/uuid"
)

type AmbulanceServiceRequest struct {
	ID              uuid.UUID                         `json:"id" gorm:"primaryKey"`
	SubmittedBy     uuid.UUID                         `json:"submittedBy" gorm:"not null"`
	AmbulanceID     *uuid.UUID                        `json:"ambulanceId"`
	SubmitterName   string                            `json:"submitterName"`
	SubmitterPhone  string                            `json:"submitterPhone"`
	SubmitterIDCard string                            `json:"submitterIdCard"`
	PatientName     string                            `json:"patientName"`
	PatientAddress  string                            `json:"patientAddress"`
	PatientAge      int                               `json:"patientAge"`
	IsInfectious    bool                              `json:"isInfectious"`
	Disease         string                            `json:"disease"`
	IsAbleToSit     bool                              `json:"isAbleToSit"`
	PickupDate      time.Time                         `json:"pickupDate"`
	PickupTime      time.Time                         `json:"pickupTime"`
	Destination     string                            `json:"destination"`
	Note            string                            `json:"note"`
	Status          Status                            `json:"status"`
	ServiceCategory ambulance_history.ServiceCategory `json:"serviceCategory"`
	RejectionReason string                            `json:"rejectionReason"`
	CreatedAt       time.Time                         `json:"createdAt"`
	UpdatedAt       time.Time                         `json:"updatedAt"`

	Ambulance *ambulance.Ambulance `json:"ambulance" gorm:"foreignKey:AmbulanceID"`
	Account   account.Account      `json:"account" gorm:"foreignKey:SubmittedBy"`
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
	StatusInService Status = "in_service"
	StatusDone      Status = "done"
)
