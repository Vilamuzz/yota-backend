package social_program_subscription

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateSocialProgramSubscriptionRequest struct {
	SocialProgramID string  `json:"socialProgramId"`
	Amount          float64 `json:"amount"`
}

type UpdateSocialProgramSubscriptionRequest struct {
	Status Status `json:"status"`
}

type SocialProgramSubscriptionQueryParams struct {
	SocialProgramID string `form:"socialProgramId"`
	AccountID       string `form:"accountId"`
	Status          string `form:"status"`
	pkg.PaginationParams
}
