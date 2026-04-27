package media

import "github.com/google/uuid"

type MediaRequest struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`

	Type    MediaType `json:"type"`
	AltText string    `json:"altText"`
	Order   int       `json:"order"`
}

type MediaMetadata struct {
	ID      string `json:"id"`
	AltText string `json:"altText"`
	Order   int    `json:"order"`
}
