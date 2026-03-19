package ambulance

import "time"

type Ambulance struct {
	ID          int       `json:"id"`
	ImageURL    string    `json:"image_url"`
	PlateNumber string    `json:"plate_number"`
	Phone       string    `json:"phone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
