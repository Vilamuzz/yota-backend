package gallery

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type GalleryResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    Category  `json:"category"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Status      Status    `json:"status"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GalleryListResponse struct {
	Galleries  []GalleryResponse    `json:"galleries"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
