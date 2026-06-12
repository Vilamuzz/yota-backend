package news

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
)

type NewsCreateRequest struct {
	Title      string                  `json:"title" form:"title"`
	Category   media.MediaCategory     `json:"category" form:"category"`
	Content    string                  `json:"content" form:"content"`
	Status     media.MediaStatus       `json:"status" form:"status"`
	CoverImage *multipart.FileHeader   `form:"coverImage" swaggerignore:"true"`
	MediaFiles []*multipart.FileHeader `form:"mediaFiles[]"`
	MediaAlts  []string                `form:"mediaAlts[]"`
}

type NewsUpdateRequest struct {
	Title             string                  `json:"title" form:"title"`
	Category          media.MediaCategory     `json:"category" form:"category"`
	Content           string                  `json:"content" form:"content"`
	Status            media.MediaStatus       `json:"status" form:"status"`
	CoverImage        *multipart.FileHeader   `form:"coverImage" swaggerignore:"true"`
	MediaFiles        []*multipart.FileHeader `form:"mediaFiles[]"`
	MediaAlts         []string                `form:"mediaAlts[]"`
	MediaOrders       []int                   `form:"mediaOrders[]"`
	MediaIDs          []string                `form:"mediaIds[]"`
	UpdateMediaAlts   []string                `form:"updateMediaAlts[]"`
	UpdateMediaOrders []int                   `form:"updateMediaOrders[]"`
}

type NewsQueryParams struct {
	Search   string              `form:"search"`
	Category media.MediaCategory `form:"category"`
	Status   media.MediaStatus   `form:"status"`
	SortBy   string              `form:"sortBy"`
	Page     int                 `form:"page"`
	Limit    int                 `form:"limit"`
}
