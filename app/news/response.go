package news

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Category    Category   `json:"category"`
	Content     string     `json:"content"`
	Image       string     `json:"image,omitempty"`
	Status      Status     `json:"status"`
	Views       int        `json:"views"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type NewsListResponse struct {
	News       []NewsResponse       `json:"news"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
