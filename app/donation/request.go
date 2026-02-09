package donation

import "time"

type DonationRequest struct {
	Title       string    `json:"title" form:"title" binding:"required,min=3,max=200"`
	Description string    `json:"description" form:"description" binding:"required,min=10,max=2000"`
	Image       string    `json:"image" form:"image"` // Optional if uploading file
	Category    Category  `json:"category" form:"category" binding:"required,oneof=education health environment"`
	FundTarget  float64   `json:"fund_target" form:"fund_target" binding:"required,gt=0"`
	DateEnd     time.Time `json:"date_end" form:"date_end" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

type UpdateDonationRequest struct {
	Title       string    `json:"title" form:"title" binding:"omitempty,min=3,max=200"`
	Description string    `json:"description" form:"description" binding:"omitempty,min=10,max=2000"`
	Image       string    `json:"image" form:"image"`
	Category    Category  `json:"category" form:"category" binding:"omitempty,oneof=education health environment"`
	FundTarget  float64   `json:"fund_target" form:"fund_target" binding:"omitempty,gt=0"`
	Status      Status    `json:"status" form:"status" binding:"omitempty,oneof=active inactive completed"`
	DateEnd     time.Time `json:"date_end" form:"date_end" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
}

type DonationQueryParams struct {
	Category string `form:"category" binding:"omitempty,oneof=education health environment"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive completed"`
	Cursor   string `form:"cursor" binding:"omitempty"` // Cursor for pagination
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
