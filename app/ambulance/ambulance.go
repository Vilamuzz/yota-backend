package ambulance

import (
	"time"

	"github.com/google/uuid"
)

type Ambulance struct {
	ID           uuid.UUID       `json:"id" gorm:"primary_key"`
	PlateNumber  string          `json:"plate_number" gorm:"unique;not null"`
	DriverName   string          `json:"driver_name" gorm:"not null"`
	DriverPhone  string          `json:"driver_phone" gorm:"not null"`
	Status       AmbulanceStatus `json:"status" gorm:"type:varchar(20);not null;default:'available'"`
	CurrentLat   float64         `json:"current_lat"`
	CurrentLng   float64         `json:"current_lng"`
	LastUpdateAt time.Time       `json:"last_update_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type AmbulanceStatus string

const (
	StatusAvailable   AmbulanceStatus = "available"
	StatusOnDuty      AmbulanceStatus = "on_duty"
	StatusOffline     AmbulanceStatus = "offline"
	StatusMaintenance AmbulanceStatus = "maintenance"
)

// LocationUpdate represents real-time location data from mobile app
type LocationUpdate struct {
	AmbulanceID string  `json:"ambulance_id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Speed       float64 `json:"speed,omitempty"`
	Heading     float64 `json:"heading,omitempty"`
	Accuracy    float64 `json:"accuracy,omitempty"`
	Timestamp   int64   `json:"timestamp"`
}

// TrackingSession represents an active tracking session
type TrackingSession struct {
	ID          uuid.UUID  `json:"id" gorm:"primary_key"`
	AmbulanceID uuid.UUID  `json:"ambulance_id" gorm:"not null"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Status      string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
}

// LocationHistory stores historical location data
type LocationHistory struct {
	ID          uuid.UUID `json:"id" gorm:"primary_key"`
	AmbulanceID uuid.UUID `json:"ambulance_id" gorm:"not null;index"`
	SessionID   uuid.UUID `json:"session_id" gorm:"not null;index"`
	Latitude    float64   `json:"latitude" gorm:"not null"`
	Longitude   float64   `json:"longitude" gorm:"not null"`
	Speed       float64   `json:"speed"`
	Heading     float64   `json:"heading"`
	RecordedAt  time.Time `json:"recorded_at" gorm:"not null;index"`
}
