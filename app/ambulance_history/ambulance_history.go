package ambulance_history

import (
	"time"

	"github.com/google/uuid"
)

type AmbulanceHistory struct {
	ID              uuid.UUID       `json:"id" gorm:"primaryKey"`
	AmbulanceID     uuid.UUID       `json:"ambulanceId" gorm:"not null"`
	DriverID        uuid.UUID       `json:"driverId" gorm:"not null"`
	ServiceCategory ServiceCategory `json:"serviceCategory" gorm:"not null"`
	Note            string          `json:"note"`
	CreatedAt       time.Time       `json:"createdAt" gorm:"not null"`
}

type ServiceCategory string

const (
	SocialService    ServiceCategory = "social_service"
	MortuaryService  ServiceCategory = "mortuary_service"
	PatientService   ServiceCategory = "patient_service"
	EmergencyService ServiceCategory = "emergency_service"
	OtherService     ServiceCategory = "other_service"
)
