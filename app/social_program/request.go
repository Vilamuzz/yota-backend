package social_program

import (
	"mime/multipart"
)

type SocialProgramRequest struct {
	Title         string                `form:"title"`
	Description   string                `form:"description"`
	CoverImage    *multipart.FileHeader `form:"coverImage" swaggerignore:"true"`
	MinimumAmount float64               `form:"minimumAmount"`
	BillingDay    int                   `form:"billingDay"`
}

type RejectSocialProgramRequest struct {
	Reason string `json:"reason"`
}

type SocialProgramQueryParams struct {
	Search       string `form:"search"`
	Status       string `form:"status"`
	IsSubscribed *bool  `form:"isSubscribed"`
	StartDate    string `form:"startDate"`
	EndDate      string `form:"endDate"`
	SortBy       string `form:"sortBy"` // e.g. "title asc", "minimum_amount desc", "billing_day asc", "total_subscribers desc", "created_at desc"
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}
