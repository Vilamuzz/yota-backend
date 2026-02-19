package gallery

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
)

type GalleryRequest struct {
	Title       string                  `json:"title" form:"title" binding:"required,min=3,max=200"`
	CategoryID  int8                    `json:"category_id" form:"category_id" binding:"required"`
	Description string                  `json:"description" form:"description" binding:"required,min=10,max=1000"`
	Published   *bool                   `json:"published" form:"published" binding:"required"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files"`
}

type UpdateGalleryRequest struct {
	Title       string                  `json:"title" form:"title" binding:"omitempty,min=3,max=200"`
	CategoryID  int8                    `json:"category_id" form:"category_id" binding:"omitempty"`
	Description string                  `json:"description" form:"description" binding:"omitempty,min=10,max=1000"`
	Published   *bool                   `json:"published" form:"published" binding:"omitempty"`
	Metadata    []media.MediaMetadata   `form:"metadata"`
	Files       []*multipart.FileHeader `form:"files"`
}

type GalleryQueryParams struct {
	CategoryID int8   `form:"category_id" binding:"omitempty"`
	Cursor     string `form:"cursor" binding:"omitempty"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
