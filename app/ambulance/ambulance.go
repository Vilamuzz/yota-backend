package ambulance

import (
	"time"

	"github.com/google/uuid"
)

type Ambulance struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey"`
	DriverID    uuid.UUID `json:"driverId"`
	Image       string    `json:"image" gorm:"not null"`
	PlateNumber string    `json:"plateNumber" gorm:"not null"`
	Phone       string    `json:"phone" gorm:"not null"`
	CreatedAt   time.Time `json:"createdAt" gorm:"not null"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"not null"`
	DeletedAt   time.Time `json:"deletedAt" gorm:"not null"`
}
