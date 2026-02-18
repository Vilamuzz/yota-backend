package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type PublishedGalleryResponse struct {
	ID          string                         `json:"id"`
	Title       string                         `json:"title"`
	Slug        string                         `json:"slug"`
	Category    string                         `json:"category"`
	Description string                         `json:"description"`
	Media       []media.PublishedMediaResponse `json:"media"`
	Views       int                            `json:"views"`
	PublishedAt string                         `json:"published_at"`
}

type GalleryResponse struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	CategoryID  int8                  `json:"category_id"`
	Description string                `json:"description"`
	Media       []media.MediaResponse `json:"media"`
	Views       int                   `json:"views"`
	PublishedAt *time.Time            `json:"published_at"`
}

type PublishedGalleryListResponseItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnail_url"`
	Views        int    `json:"views"`
	PublishedAt  string `json:"published_at"`
}

type PublishedGalleryListResponse struct {
	Galleries  []PublishedGalleryListResponseItem `json:"galleries"`
	Pagination pkg.CursorPagination               `json:"pagination"`
}
