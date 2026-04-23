package gallery

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type GalleryRequest struct {
	Title       string                  `json:"title" form:"title"`
	Category    media.MediaCategory     `json:"category" form:"category"`
	Description string                  `json:"description" form:"description"`
	Published   *bool                   `json:"published" form:"published"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files" swaggerignore:"true"`
}

type UpdateGalleryRequest struct {
	Title       string                  `json:"title" form:"title"`
	Category    media.MediaCategory     `json:"category" form:"category"`
	Description string                  `json:"description" form:"description"`
	Published   *bool                   `json:"published" form:"published"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files" swaggerignore:"true"`
}

type GalleryQueryParams struct {
	Category media.MediaCategory `form:"category"`
	pkg.CursorPagination
}
