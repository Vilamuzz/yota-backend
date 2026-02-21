package donation

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type PublishedDonationResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Category    Category  `json:"category"`
	FundTarget  float64   `json:"fund_target"`
	Status      Status    `json:"status"`
	DateEnd     time.Time `json:"date_end"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DonationResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Category    Category  `json:"category"`
	FundTarget  float64   `json:"fund_target"`
	Status      Status    `json:"status"`
	DateEnd     time.Time `json:"date_end"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DonationListResponse struct {
	Donations  []DonationResponse   `json:"donations"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
