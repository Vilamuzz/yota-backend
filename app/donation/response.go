package donation

import "time"

type DonationResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       string    `json:"image,omitempty"`
	Category    Category  `json:"category"`
	FundTarget  float64   `json:"fund_target"`
	Status      Status    `json:"status"`
	DateEnd     time.Time `json:"date_end"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DonationListResponse struct {
	Donations  []DonationResponse `json:"donations"`
	Pagination CursorPagination   `json:"pagination"`
}

type CursorPagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	Limit      int    `json:"limit"`
}
