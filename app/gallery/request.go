package gallery

import "github.com/Vilamuzz/yota-backend/app/media"

type GalleryRequest struct {
	Title       string               `json:"title" form:"title" binding:"required,min=3,max=200"`
	Category    Category             `json:"category" form:"category" binding:"required,oneof=photography painting sculpture digital mixed"`
	Description string               `json:"description" form:"description" binding:"required,min=10,max=1000"`
	Published   bool                 `json:"published" form:"published" binding:"required"`
	Media       []media.MediaRequest `json:"media" binding:"omitempty,dive"`
}

type UpdateGalleryRequest struct {
	Title       string               `json:"title" form:"title" binding:"omitempty,min=3,max=200"`
	Category    Category             `json:"category" form:"category" binding:"omitempty,oneof=photography painting sculpture digital mixed"`
	Description string               `json:"description" form:"description" binding:"omitempty,min=10,max=1000"`
	Published   bool                 `json:"published" form:"published" binding:"omitempty"`
	Media       []media.MediaRequest `json:"media" binding:"omitempty,dive"`
}

type GalleryQueryParams struct {
	Category Category `form:"category" binding:"omitempty,oneof=photography painting sculpture digital mixed"`
	Cursor   string   `form:"cursor" binding:"omitempty"`
	Limit    int      `form:"limit" binding:"omitempty,min=1,max=100"`
}
