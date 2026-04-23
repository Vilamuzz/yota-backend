package ambulance

import (
	"time"

	"github.com/google/uuid"
)

type Ambulance struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey"`
	DriverID    uuid.UUID `json:"driver_id"`
	Image       string    `json:"image" gorm:"not null"`
	PlateNumber string    `json:"plate_number" gorm:"not null"`
	Phone       string    `json:"phone" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`
	DeletedAt   time.Time `json:"deleted_at" gorm:"not null"`
}
