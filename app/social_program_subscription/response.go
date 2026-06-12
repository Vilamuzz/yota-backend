package social_program_subscription

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramSubscriptionResponse struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	Status           string    `json:"status"`
	TotalPaidPeriods int       `json:"totalPaidPeriods"`
	TotalDonation    float64   `json:"totalDonation"`
	CreatedAt        time.Time `json:"createdAt"`
}

type SubscribersResponse struct {
	ID                string  `json:"id"`
	Username          string  `json:"username"`
	Email             string  `json:"email"`
	TotalSubscription int     `json:"totalSubscription"`
	TotalDonation     float64 `json:"totalDonation"`
}

type SubscriberSubscriptionResponse struct {
	ID                 string  `json:"id"`
	SocialProgramTitle string  `json:"socialProgramTitle"`
	Status             string  `json:"status"`
	TotalPaidPeriods   int     `json:"totalPaidPeriods"`
	TotalDonation      float64 `json:"totalDonation"`
	CreatedAt          string  `json:"createdAt"`
}

type SubscriberSubscriptionListResponse struct {
	Subscriptions []SubscriberSubscriptionResponse `json:"subscriptions"`
	Pagination    pkg.CursorPagination             `json:"pagination"`
}

type SocialProgramSubscriptionListResponse struct {
	Subscriptions []SocialProgramSubscriptionResponse `json:"subscriptions"`
	Pagination    pkg.CursorPagination                `json:"pagination"`
}

type SubscriptionsListResponse struct {
	Subscribers []SubscribersResponse `json:"subscribers"`
	Pagination  pkg.CursorPagination  `json:"pagination"`
}

func (s *SocialProgramSubscription) toSocialProgramSubscriptionResponse() SocialProgramSubscriptionResponse {
	username := "Unknown"
	if s.Account != nil {
		username = s.Account.UserProfile.Username
	}

	return SocialProgramSubscriptionResponse{
		ID:               s.ID.String(),
		Username:         username,
		Status:           string(s.Status),
		TotalPaidPeriods: s.TotalPaidPeriods,
		TotalDonation:    s.TotalDonation,
		CreatedAt:        s.CreatedAt,
	}
}

func (s *SocialProgramSubscription) toSubscriberSubscriptionResponse(totalDonation float64) SubscriberSubscriptionResponse {
	programName := "Unknown"
	if s.SocialProgram != nil {
		programName = s.SocialProgram.Title
	}

	return SubscriberSubscriptionResponse{
		ID:                 s.ID.String(),
		SocialProgramTitle: programName,
		Status:             string(s.Status),
		TotalPaidPeriods:   s.TotalPaidPeriods,
		TotalDonation:      totalDonation,
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
	}
}

func toSubscriberSubscriptionListResponse(subscriptions []SocialProgramSubscription, pagination pkg.CursorPagination, donationMap map[string]float64) SubscriberSubscriptionListResponse {
	var responses []SubscriberSubscriptionResponse
	for _, sub := range subscriptions {
		donation := donationMap[sub.ID.String()]
		responses = append(responses, sub.toSubscriberSubscriptionResponse(donation))
	}
	if responses == nil {
		responses = []SubscriberSubscriptionResponse{}
	}
	return SubscriberSubscriptionListResponse{
		Subscriptions: responses,
		Pagination:    pagination,
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

func (s *SocialProgramSubscription) toSubscribersResponse(stats SubscriberStats) SubscribersResponse {
	username := "Unknown"
	email := "Unknown"
	if s.Account != nil {
		username = s.Account.UserProfile.Username
		email = s.Account.Email
	}
	return SubscribersResponse{
		ID:                s.AccountID.String(),
		Username:          username,
		Email:             email,
		TotalSubscription: stats.TotalSubscription,
		TotalDonation:     stats.TotalDonation,
	}
}

func toSubscriptionsListResponse(subscriptions []SocialProgramSubscription, pagination pkg.CursorPagination, statsMap map[string]SubscriberStats) SubscriptionsListResponse {
	var responses []SubscribersResponse
	for _, sub := range subscriptions {
		stats := statsMap[sub.AccountID.String()]
		responses = append(responses, sub.toSubscribersResponse(stats))
	}
	if responses == nil {
		responses = []SubscribersResponse{}
	}
	return SubscriptionsListResponse{
		Subscribers: responses,
		Pagination:  pagination,
	}
}
