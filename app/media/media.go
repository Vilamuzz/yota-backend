package media

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	NewsID    uuid.UUID `json:"news_id" gorm:"index"`
	GalleryID uuid.UUID `json:"gallery_id" gorm:"index"`
	Type      MediaType `json:"type" gorm:"not null"`

	URL       string    `json:"url" gorm:"not null"`
	AltText   string    `json:"alt_text"`
	Order     int       `json:"order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
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
	SosialEvent MediaCategory = "kegiatan_sosial"
	Disaster    MediaCategory = "bencana_alam"
	Health      MediaCategory = "kesehatan"
	Environment MediaCategory = "lingkungan"
	Others      MediaCategory = "lainnya"
)
