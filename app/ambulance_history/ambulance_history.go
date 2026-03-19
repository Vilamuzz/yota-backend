package ambulance_history

import "time"

type AmbulanceHistory struct {
	ID              int             `json:"id"`
	AmbulanceID     int             `json:"ambulance_id"`
	UserID          int             `json:"user_id"`
	ServiceCategory ServiceCategory `json:"service_category"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ServiceCategory string

const (
	SocialService    ServiceCategory = "social_service"
	MortuaryService  ServiceCategory = "mortuary_service"
	PatientService   ServiceCategory = "patient_service"
	EmergencyService ServiceCategory = "emergency_service"
	OtherService     ServiceCategory = "other_service"
)
