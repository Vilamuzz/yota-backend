package social_program

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramRequest struct {
	Title         string                `form:"title"`
	Description   string                `form:"description"`
	CoverImage    *multipart.FileHeader `form:"cover_image" swaggerignore:"true"`
	Status        Status                `form:"status"`
	MinimumAmount float64               `form:"minimum_amount"`
	BillingDay    int                   `form:"billing_day"`
}

type SocialProgramQueryParams struct {
	Search       string `form:"search"`
	Status       string `form:"status"`
	IsSubscribed *bool  `form:"is_subscribed"`
	pkg.PaginationParams
}
