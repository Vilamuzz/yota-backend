package donation_program

import (
	"mime/multipart"
)

type DonationProgramRequest struct {
	Title       string                `json:"title" form:"title"`
	CoverImage  *multipart.FileHeader `json:"coverImage" form:"coverImage" swaggerignore:"true"`
	Category    Category              `json:"category" form:"category"`
	Description string                `json:"description" form:"description"`
	FundTarget  float64               `json:"fundTarget" form:"fundTarget"`
	Status      Status                `json:"status" form:"status"`
	StartDate   string                `json:"startDate" form:"startDate"`
	EndDate     string                `json:"endDate" form:"endDate"`
}

type DonationProgramQueryParams struct {
	Search   string   `form:"search"`
	Category Category `form:"category"`
	Status   string   `form:"status"`
	SortBy   string   `form:"sortBy"`
	Page     int      `form:"page"`
	Limit    int      `form:"limit"`
}
