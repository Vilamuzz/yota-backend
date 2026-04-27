package social_program_subscription

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramSubscriptionResponse struct {
	ID              string    `json:"id"`
	SocialProgramID string    `json:"socialProgramId"`
	AccountID       string    `json:"accountId"`
	Status          string    `json:"status"`
	Amount          float64   `json:"amount"`
	CreatedAt       time.Time `json:"createdAt"`
}

type SocialProgramSubscriptionListResponse struct {
	SocialProgramSubscriptions []SocialProgramSubscriptionResponse `json:"socialProgramSubscriptions"`
	Pagination                 pkg.CursorPagination                `json:"pagination"`
}

func (s *SocialProgramSubscription) toSocialProgramSubscriptionResponse() SocialProgramSubscriptionResponse {
	return SocialProgramSubscriptionResponse{
		ID:              s.ID.String(),
		SocialProgramID: s.SocialProgramID.String(),
		AccountID:       s.AccountID.String(),
		Status:          string(s.Status),
		Amount:          s.Amount,
		CreatedAt:       s.CreatedAt,
	}
}

func toSocialProgramSubscriptionListResponse(subscriptions []SocialProgramSubscription, pagination pkg.CursorPagination) SocialProgramSubscriptionListResponse {
	var responses []SocialProgramSubscriptionResponse
	for _, sub := range subscriptions {
		responses = append(responses, sub.toSocialProgramSubscriptionResponse())
	}
	if responses == nil {
		responses = []SocialProgramSubscriptionResponse{}
	}
	return SocialProgramSubscriptionListResponse{
		SocialProgramSubscriptions: responses,
		Pagination:                 pagination,
	}
}
