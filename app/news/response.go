package news

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsResponse struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	CoverImage  string                `json:"coverImage"`
	Category    media.MediaCategory   `json:"category"`
	Content     string                `json:"content"`
	Media       []media.MediaResponse `json:"media"`
	Views       int                   `json:"views"`
	Status      media.MediaStatus     `json:"status"`
	PublishedAt *time.Time            `json:"publishedAt"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

type NewsListResponseItem struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Slug        string              `json:"slug"`
	CoverImage  string              `json:"coverImage"`
	Category    media.MediaCategory `json:"category"`
	Views       int                 `json:"views"`
	Status      media.MediaStatus   `json:"status"`
	PublishedAt *time.Time          `json:"publishedAt"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type NewsListResponse struct {
	News       []NewsListResponseItem `json:"news"`
	Pagination pkg.CursorPagination   `json:"pagination"`
}

func (n *News) toNewsResponse() NewsResponse {
	mediaResponses := make([]media.MediaResponse, 0, len(n.Media))
	for _, m := range n.Media {
		mediaResponses = append(mediaResponses, media.MediaResponse{
			ID:      m.ID.String(),
			NewsID:  m.NewsID.String(),
			Type:    string(m.Type),
			URL:     m.URL,
			Alt:     m.Alt,
			Order:   m.Order,
		})
	}

	return NewsResponse{
		ID:          n.ID.String(),
		Title:       n.Title,
		Slug:        n.Slug,
		CoverImage:  n.CoverImage,
		Category:    n.Category,
		Content:     n.Content,
		Media:       mediaResponses,
		Views:       n.Views,
		Status:      n.Status,
		PublishedAt: n.PublishedAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

func (n *News) toNewsListResponseItem() NewsListResponseItem {
	return NewsListResponseItem{
		ID:          n.ID.String(),
		Title:       n.Title,
		Slug:        n.Slug,
		CoverImage:  n.CoverImage,
		Category:    n.Category,
		Views:       n.Views,
		Status:      n.Status,
		PublishedAt: n.PublishedAt,
		CreatedAt:   n.CreatedAt,
	}
}

func toNewsListResponse(newsList []News, pagination pkg.CursorPagination) NewsListResponse {
	var newsResponses []NewsListResponseItem
	for _, n := range newsList {
		newsResponses = append(newsResponses, n.toNewsListResponseItem())
	}

	if newsResponses == nil {
		newsResponses = []NewsListResponseItem{}
	}

	return NewsListResponse{
		News:       newsResponses,
		Pagination: pagination,
	}
}
