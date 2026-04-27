package news

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Category    media.MediaCategory `json:"category"`
	Content     string              `json:"content"`
	Image       string              `json:"image"`
	Status      media.MediaStatus   `json:"status"`
	Views       int                 `json:"views"`
	PublishedAt *time.Time          `json:"publishedAt"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type NewsListResponse struct {
	News       []NewsResponse       `json:"news"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
