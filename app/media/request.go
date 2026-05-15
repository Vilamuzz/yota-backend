package media

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type MediaRequest struct {
	ID   uuid.UUID             `form:"id"`
	File *multipart.FileHeader `form:"media[][file]"`
	Alt  string                `form:"media[][alt]"`
}

type MediaMetadata struct {
	ID    string `json:"id"`
	Alt   string `json:"alt"`
	Order int    `json:"order"`
}
