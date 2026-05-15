package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type GalleryResponse struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	CoverImage  string                `json:"coverImage"`
	Category    media.MediaCategory   `json:"category"`
	Description string                `json:"description"`
	Media       []media.MediaResponse `json:"media"`
	Views       int                   `json:"views"`
	Status      media.MediaStatus     `json:"status"`
	CreatedAt   time.Time             `json:"createdAt"`
}

type GalleryListResponseItem struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	CoverImage  string              `json:"coverImage"`
	Slug        string              `json:"slug"`
	Category    media.MediaCategory `json:"category"`
	Description string              `json:"description"`
	Views       int                 `json:"views"`
	Status      media.MediaStatus   `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type GalleryListResponse struct {
	Galleries  []GalleryListResponseItem `json:"galleries"`
	Pagination pkg.CursorPagination      `json:"pagination"`
}

func (g *Gallery) toGalleryResponse() GalleryResponse {
	mediaResponses := make([]media.MediaResponse, 0, len(g.Media))
	for _, m := range g.Media {
		mediaResponses = append(mediaResponses, media.MediaResponse{
			ID:        m.ID.String(),
			GalleryID: m.GalleryID.String(),
			Type:      string(m.Type),
			URL:       m.URL,
			Alt:       m.Alt,
			Order:     m.Order,
		})
	}

	return GalleryResponse{
		ID:          g.ID.String(),
		Title:       g.Title,
		Slug:        g.Slug,
		CoverImage:  g.CoverImage,
		Category:    g.Category,
		Description: g.Description,
		Media:       mediaResponses,
		Views:       g.Views,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
	}
}

func (g *Gallery) toGalleryListResponseItem() GalleryListResponseItem {
	return GalleryListResponseItem{
		ID:          g.ID.String(),
		Title:       g.Title,
		Slug:        g.Slug,
		CoverImage:  g.CoverImage,
		Category:    g.Category,
		Description: g.Description,
		Views:       g.Views,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
	}
}

func toGalleryListResponse(galleries []Gallery, pagination pkg.CursorPagination) GalleryListResponse {
	var galleryResponses []GalleryListResponseItem
	for _, g := range galleries {
		galleryResponses = append(galleryResponses, g.toGalleryListResponseItem())
	}

	if galleryResponses == nil {
		galleryResponses = []GalleryListResponseItem{}
	}

	return GalleryListResponse{
		Galleries:  galleryResponses,
		Pagination: pagination,
	}
}
