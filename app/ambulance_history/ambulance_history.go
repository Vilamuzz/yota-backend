package ambulance_history

import (
	"time"

	"github.com/google/uuid"
)

type AmbulanceHistory struct {
	ID              uuid.UUID       `json:"id" gorm:"primaryKey"`
	AmbulanceID     uuid.UUID       `json:"ambulance_id" gorm:"not null"`
	DriverID        uuid.UUID       `json:"driver_id" gorm:"not null"`
	ServiceCategory ServiceCategory `json:"service_category" gorm:"not null"`
	Note            string          `json:"note"`
	CreatedAt       time.Time       `json:"created_at" gorm:"not null"`
}

type ServiceCategory string

const (
	SocialService    ServiceCategory = "social_service"
	MortuaryService  ServiceCategory = "mortuary_service"
	PatientService   ServiceCategory = "patient_service"
	EmergencyService ServiceCategory = "emergency_service"
	OtherService     ServiceCategory = "other_service"
)
