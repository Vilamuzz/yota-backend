package social_program_subscription

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateSocialProgramSubscriptionRequest struct {
	SocialProgramID string  `json:"social_program_id"`
	Amount          float64 `json:"amount"`
}

type UpdateSocialProgramSubscriptionRequest struct {
	Status Status `json:"status"`
}

type SocialProgramSubscriptionQueryParams struct {
	SocialProgramID string `form:"social_program_id"`
	AccountID       string `form:"account_id"`
	Status          string `form:"status"`
	pkg.PaginationParams
}
