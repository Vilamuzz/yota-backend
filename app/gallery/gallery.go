package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
)

type Gallery struct {
	ID          string     `json:"id" gorm:"primary_key"`
	Title       string     `json:"title" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"unique;not null"`
	CategoryID  int8       `json:"category_id" gorm:"not null"`
	Description string     `json:"description" gorm:"not null"`
	Views       int        `json:"views" gorm:"not null;default:0"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`

	Media         []media.Media       `gorm:"polymorphic:Entity;"`
	CategoryMedia media.CategoryMedia `gorm:"foreignKey:CategoryID"`
}
