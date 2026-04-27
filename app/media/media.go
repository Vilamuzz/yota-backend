package media

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	NewsID    uuid.UUID `json:"newsId" gorm:"index"`
	GalleryID uuid.UUID `json:"galleryId" gorm:"index"`
	Type      MediaType `json:"type" gorm:"not null"`

	URL       string    `json:"url" gorm:"not null"`
	AltText   string    `json:"altText"`
	Order     int       `json:"order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null"`
}

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

type MediaStatus string

const (
	MediaStatusDraft     MediaStatus = "draft"
	MediaStatusPublished MediaStatus = "published"
	MediaStatusArchived  MediaStatus = "archived"
)

type MediaCategory string

const (
	SocialEvent MediaCategory = "Kegiatan Sosial"
	Disaster    MediaCategory = "Bencana Alam"
	Health      MediaCategory = "Kesehatan"
	Environment MediaCategory = "Lingkungan"
	Others      MediaCategory = "Lainnya"
)
