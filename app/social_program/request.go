package social_program

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramRequest struct {
	Title         string                `form:"title"`
	Description   string                `form:"description"`
	CoverImage    *multipart.FileHeader `form:"coverImage" swaggerignore:"true"`
	Status        Status                `form:"status"`
	MinimumAmount float64               `form:"minimumAmount"`
	BillingDay    int                   `form:"billingDay"`
}

type SocialProgramQueryParams struct {
	Search       string `form:"search"`
	Status       string `form:"status"`
	IsSubscribed *bool  `form:"isSubscribed"`
	pkg.PaginationParams
}
