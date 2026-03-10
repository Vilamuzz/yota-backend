package donation

import (
	"mime/multipart"
	"time"
)

type DonationRequest struct {
	Title       string                `json:"title" form:"title"`
	Description string                `json:"description" form:"description"`
	Image       *multipart.FileHeader `json:"image" form:"image"`
	Category    Category              `json:"category" form:"category"`
	Status      bool                  `json:"status" form:"status"`
	FundTarget  float64               `json:"fund_target" form:"fund_target"`
	DateEnd     time.Time             `json:"date_end" form:"date_end" time_format:"2006-01-02"`
}

type UpdateDonationRequest struct {
	Title       string                `json:"title" form:"title"`
	Description string                `json:"description" form:"description"`
	Image       *multipart.FileHeader `json:"image" form:"image"`
	Category    Category              `json:"category" form:"category"`
	FundTarget  float64               `json:"fund_target" form:"fund_target"`
	Status      *bool                 `json:"status" form:"status"`
	DateEnd     time.Time             `json:"date_end" form:"date_end" time_format:"2006-01-02"`
}

type DonationQueryParams struct {
	Search   string `form:"search"`
	Category string `form:"category"`
	Status   string `form:"status"`
	Cursor   string `form:"cursor"`
	Limit    int    `form:"limit"`
}
