package media

import "github.com/google/uuid"

type MediaRequest struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`

	Type    MediaType `json:"type"`
	AltText string    `json:"alt_text"`
	Order   int       `json:"order"`
}

type MediaMetadata struct {
	ID      string `json:"id"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order"`
}
