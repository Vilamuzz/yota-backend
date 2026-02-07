package donation

import "time"

type DonationRequest struct {
	Title       string    `json:"title" binding:"required,min=3,max=200"`
	Description string    `json:"description" binding:"required,min=10,max=2000"`
	Image       string    `json:"image" binding:"omitempty,url"`
	Category    Category  `json:"category" binding:"required,oneof=education health environment"`
	FundTarget  float64   `json:"fund_target" binding:"required,gt=0"`
	DateEnd     time.Time `json:"date_end" binding:"required"`
}

type UpdateDonationRequest struct {
	Title       string    `json:"title" binding:"omitempty,min=3,max=200"`
	Description string    `json:"description" binding:"omitempty,min=10,max=2000"`
	Image       string    `json:"image" binding:"omitempty,url"`
	Category    Category  `json:"category" binding:"omitempty,oneof=education health environment"`
	FundTarget  float64   `json:"fund_target" binding:"omitempty,gt=0"`
	Status      Status    `json:"status" binding:"omitempty,oneof=active inactive completed"`
	DateEnd     time.Time `json:"date_end" binding:"omitempty"`
}

type DonationQueryParams struct {
	Category string `form:"category" binding:"omitempty,oneof=education health environment"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive completed"`
	Cursor   string `form:"cursor" binding:"omitempty"` // Cursor for pagination
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
