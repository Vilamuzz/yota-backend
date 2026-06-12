package social_program_subscription

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateSocialProgramSubscriptionOfflineRequest struct {
	AccountID string `json:"accountId"`
}

type UpdateSocialProgramSubscriptionRequest struct {
	Status Status `json:"status"`
}

type SocialProgramSubscriptionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
