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
	CategoryID  int8                  `json:"category_id"`
	Description string                `json:"description"`
	Media       []media.MediaResponse `json:"media"`
	Views       int                   `json:"views"`
	PublishedAt *time.Time            `json:"published_at"`
}

type GalleryListResponseItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	CategoryID  int8       `json:"category_id"`
	Description string     `json:"description"`
	Views       int        `json:"views"`
	PublishedAt *time.Time `json:"published_at"`
}

type GalleryListResponse struct {
	Galleries  []GalleryListResponseItem `json:"galleries"`
	Pagination pkg.CursorPagination      `json:"pagination"`
}

func (g *Gallery) toGalleryResponse() GalleryResponse {
	mediaResponses := make([]media.MediaResponse, 0, len(g.Media))
	for _, m := range g.Media {
		mediaResponses = append(mediaResponses, media.MediaResponse{
			ID:         m.ID,
			EntityID:   m.EntityID,
			EntityType: m.EntityType,
			Type:       m.Type,
			URL:        m.URL,
			AltText:    m.AltText,
			Order:      m.Order,
		})
	}

	return GalleryResponse{
		ID:          g.ID,
		Title:       g.Title,
		Slug:        g.Slug,
		CategoryID:  g.CategoryID,
		Description: g.Description,
		Media:       mediaResponses,
		Views:       g.Views,
		PublishedAt: g.PublishedAt,
	}
}

func (g *Gallery) toGalleryListResponseItem() GalleryListResponseItem {
	return GalleryListResponseItem{
		ID:          g.ID,
		Title:       g.Title,
		Slug:        g.Slug,
		CategoryID:  g.CategoryID,
		Description: g.Description,
		Views:       g.Views,
		PublishedAt: g.PublishedAt,
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

func (g *Gallery) toPublishedGalleryResponse() PublishedGalleryResponse {
	mediaResponses := make([]media.PublishedMediaResponse, 0, len(g.Media))
	for _, m := range g.Media {
		mediaResponses = append(mediaResponses, media.PublishedMediaResponse{
			URL:     m.URL,
			AltText: m.AltText,
			Order:   m.Order,
		})
	}

	publishedAt := ""
	if g.PublishedAt != nil {
		publishedAt = g.PublishedAt.Format(time.RFC3339)
	}

	return PublishedGalleryResponse{
		ID:          g.ID,
		Title:       g.Title,
		Slug:        g.Slug,
		Category:    g.CategoryMedia.Category,
		Description: g.Description,
		Media:       mediaResponses,
		Views:       g.Views,
		PublishedAt: publishedAt,
	}
}

func (g *Gallery) toPublishedGalleryListResponseItem() PublishedGalleryListResponseItem {
	thumbnailURL := ""
	if len(g.Media) > 0 {
		lowestOrderMedia := g.Media[0]
		for _, m := range g.Media {
			if m.Order < lowestOrderMedia.Order {
				lowestOrderMedia = m
			}
		}
		thumbnailURL = lowestOrderMedia.URL
	}

	return PublishedGalleryListResponseItem{
		ID:           g.ID,
		Title:        g.Title,
		Slug:         g.Slug,
		Category:     g.CategoryMedia.Category,
		Description:  g.Description,
		ThumbnailURL: thumbnailURL,
		Views:        g.Views,
		PublishedAt:  g.PublishedAt.Format(time.RFC3339),
	}
}

func toPublishedGalleryListResponse(galleries []Gallery, pagination pkg.CursorPagination) PublishedGalleryListResponse {
	var responses []PublishedGalleryListResponseItem
	for _, g := range galleries {
		responses = append(responses, g.toPublishedGalleryListResponseItem())
	}
	if responses == nil {
		responses = []PublishedGalleryListResponseItem{}
	}
	return PublishedGalleryListResponse{
		Galleries:  responses,
		Pagination: pagination,
	}
}
