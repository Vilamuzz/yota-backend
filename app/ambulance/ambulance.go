package ambulance

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type Ambulance struct {
	ID          uuid.UUID       `json:"id" gorm:"primaryKey"`
	DriverID    uuid.UUID       `json:"driverId"`
	Image       string          `json:"image"`
	PlateNumber string          `json:"plateNumber"`
	Status      AmbulanceStatus `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *time.Time      `json:"deletedAt"`

	Driver account.Account `json:"driver" gorm:"foreignKey:DriverID"`
}

type AmbulanceStatus string

const (
	AmbulanceStatusAvailable   AmbulanceStatus = "available"
	AmbulanceStatusInUse       AmbulanceStatus = "in use"
	AmbulanceStatusMaintenance AmbulanceStatus = "maintenance"
)

func (s AmbulanceStatus) IsValid() bool {
	switch s {
	case AmbulanceStatusAvailable, AmbulanceStatusInUse, AmbulanceStatusMaintenance:
		return true
	}
	return false
}
