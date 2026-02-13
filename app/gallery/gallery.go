package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
)

type Gallery struct {
	ID          string        `json:"id" gorm:"primary_key"`
	Title       string        `json:"title" gorm:"not null"`
	Slug        string        `json:"slug" gorm:"unique;not null"`
	Category    Category      `json:"category" gorm:"not null"`
	Description string        `json:"description" gorm:"not null"`
	Views       int           `json:"views" gorm:"not null;default:0"`
	Media       []media.Media `json:"media" gorm:"polymorphic:Entity;"`
	PublishedAt *time.Time    `json:"published_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   time.Time     `json:"deleted_at" gorm:"index"`
}

type Category string

const (
	CategoryPhotography Category = "photography"
	CategoryPainting    Category = "painting"
	CategorySculpture   Category = "sculpture"
	CategoryDigital     Category = "digital"
	CategoryMixed       Category = "mixed"
)
