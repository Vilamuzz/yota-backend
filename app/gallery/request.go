package gallery

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type GalleryRequest struct {
	Title       string                  `json:"title" form:"title"`
	CategoryID  int8                    `json:"category_id" form:"category_id"`
	Description string                  `json:"description" form:"description"`
	Published   *bool                   `json:"published" form:"published"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files"`
}

type UpdateGalleryRequest struct {
	Title       string                  `json:"title" form:"title"`
	CategoryID  int8                    `json:"category_id" form:"category_id"`
	Description string                  `json:"description" form:"description"`
	Published   *bool                   `json:"published" form:"published"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files"`
}

type GalleryQueryParams struct {
	CategoryID int8 `form:"category_id"`
	pkg.CursorPagination
}
