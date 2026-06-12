package gallery

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type GalleryCreateRequest struct {
	Title       string                  `json:"title" form:"title"`
	CoverImage  *multipart.FileHeader   `form:"coverImage" swaggerignore:"true"`
	Category    media.MediaCategory     `json:"category" form:"category"`
	Description string                  `json:"description" form:"description"`
	Status      media.MediaStatus       `json:"status" form:"status"`
	MediaFiles  []*multipart.FileHeader `form:"mediaFiles[]"`
	MediaAlts   []string                `form:"mediaAlts[]"`
}

type GalleryUpdateRequest struct {
	Title             string                  `json:"title" form:"title"`
	CoverImage        *multipart.FileHeader   `form:"coverImage" swaggerignore:"true"`
	Category          media.MediaCategory     `json:"category" form:"category"`
	Description       string                  `json:"description" form:"description"`
	Status            media.MediaStatus       `json:"status" form:"status"`
	MediaFiles        []*multipart.FileHeader `form:"mediaFiles[]"`
	MediaAlts         []string                `form:"mediaAlts[]"`
	MediaOrders       []int                   `form:"mediaOrders[]"`
	MediaIDs          []string                `form:"mediaIds[]"`
	UpdateMediaAlts   []string                `form:"updateMediaAlts[]"`
	UpdateMediaOrders []int                   `form:"updateMediaOrders[]"`
}

type GalleryQueryParams struct {
	Category media.MediaCategory `form:"category"`
	Status   media.MediaStatus   `form:"status"`
	pkg.PaginationParams
}
