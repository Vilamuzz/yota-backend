package social_program

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateSocialProgramRequest struct {
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	ImageURL      multipart.FileHeader `json:"image_url"`
	Status        bool                 `json:"status"`
	MinimumAmount float64              `json:"minimum_amount"`
	BillingDay    int                  `json:"billing_day"`
}

type UpdateSocialProgramRequest struct {
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	ImageURL      multipart.FileHeader `json:"image_url"`
	Status        Status               `json:"status"`
	MinimumAmount float64              `json:"minimum_amount"`
	BillingDay    int                  `json:"billing_day"`
}

type SocialProgramQueryParams struct {
	Status       string `form:"status"`
	IsSubscribed bool   `form:"is_subscribed"`
	pkg.PaginationParams
}
