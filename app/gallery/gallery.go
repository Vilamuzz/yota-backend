package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
)

type Gallery struct {
	ID          string        `json:"id" gorm:"primary_key"`
	Title       string        `json:"title" gorm:"not null"`
	Category    Category      `json:"category" gorm:"not null"`
	Description string        `json:"description" gorm:"not null"`
	Status      Status        `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Views       int           `json:"views" gorm:"not null;default:0"`
	Media       []media.Media `json:"media" gorm:"polymorphic:Entity;"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusArchived Status = "archived"
)

type Category string

const (
	CategoryPhotography Category = "photography"
	CategoryPainting    Category = "painting"
	CategorySculpture   Category = "sculpture"
	CategoryDigital     Category = "digital"
	CategoryMixed       Category = "mixed"
)
