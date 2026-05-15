package social_program_subscription

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramSubscriptionResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type SocialProgramSubscriptionListResponse struct {
	Subscriptions []SocialProgramSubscriptionResponse `json:"subscriptions"`
	Pagination    pkg.CursorPagination                `json:"pagination"`
}

func (s *SocialProgramSubscription) toSocialProgramSubscriptionResponse() SocialProgramSubscriptionResponse {
	username := "Unknown"
	if s.Account != nil {
		username = s.Account.UserProfile.Username
	}

	return SocialProgramSubscriptionResponse{
		ID:        s.ID.String(),
		Username:  username,
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt,
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
		Subscriptions: responses,
		Pagination:    pagination,
	}
}
