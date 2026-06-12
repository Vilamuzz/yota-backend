package social_program_subscription

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateSocialProgramSubscriptionOfflineRequest struct {
	AccountID string `json:"accountId"`
}

type UpdateSocialProgramSubscriptionRequest struct {
	Status Status `json:"status"`
}

type SocialProgramSubscriptionQueryParams struct {
	Search string `form:"search"`
	Status string `form:"status"`
	SortBy string `form:"sortBy"` // e.g. "total_donation desc", "total_paid_periods desc"
	pkg.OffsetPagination
}
