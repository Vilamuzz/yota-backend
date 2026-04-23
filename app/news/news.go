package news

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/google/uuid"
)

type News struct {
	ID          uuid.UUID           `json:"id" gorm:"primaryKey"`
	Title       string              `json:"title" gorm:"not null"`
	Slug        string              `json:"slug" gorm:"not null;unique"`
	Category    media.MediaCategory `json:"category" gorm:"not null"`
	CoverImage  string              `json:"cover_image" gorm:"not null"`
	Content     string              `json:"content" gorm:"not null"`
	Status      media.MediaStatus   `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	Views       int                 `json:"views" gorm:"not null;default:0"`
	PublishedAt *time.Time          `json:"published_at" gorm:"index"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   *time.Time          `json:"deleted_at" gorm:"index;not null"`

	Media []media.Media `json:"media" gorm:"foreignKey:NewsID"`
}
