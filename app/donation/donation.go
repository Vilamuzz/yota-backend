package donation

import "time"

type Donation struct {
	ID          string    `json:"id" gorm:"primary_key"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	ImageURL    string    `json:"image_url"`
	Category    Category  `json:"category" gorm:"not null"`
	FundTarget  float64   `json:"fund_target" gorm:"not null"`
	Status      Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	DateEnd     time.Time `json:"date_end"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusDraft     Status = "draft"
	StatusCompleted Status = "completed"
)

type Category string

const (
	CategoryEducation   Category = "education"
	CategoryHealth      Category = "health"
	CategoryEnvironment Category = "environment"
)
