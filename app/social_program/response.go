package social_program

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramResponse struct {
	ID               string           `json:"id"`
	Slug             string           `json:"slug"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	CoverImage       string           `json:"coverImage"`
	Status           Status           `json:"status"`
	SubmissionStatus SubmissionStatus `json:"submissionStatus"`
	RejectionReason  string           `json:"rejectionReason"`
	IsSubscribed     bool             `json:"isSubscribed"`
	MinimumAmount    float64          `json:"minimumAmount"`
	BillingDay       int              `json:"billingDay"`
	CreatedAt        string           `json:"createdAt"`
}

type SocialProgramSubscriptionResponse struct {
	ID              string             `json:"id"`
	SocialProgramID string             `json:"socialProgramId"`
	AccountID       string             `json:"accountId"`
	Amount          float64            `json:"amount"`
	Status          SubscriptionStatus `json:"status"`
	CreatedAt       string             `json:"createdAt"`
}

type SocialProgramListResponse struct {
	SocialPrograms []SocialProgramResponse `json:"socialPrograms"`
	Pagination     pkg.CursorPagination    `json:"pagination"`
}

func (r *SocialProgram) toSocialProgramResponse() SocialProgramResponse {
	return SocialProgramResponse{
		ID:               r.ID.String(),
		Slug:             r.Slug,
		Title:            r.Title,
		Description:      r.Description,
		CoverImage:       r.CoverImage,
		Status:           r.Status,
		SubmissionStatus: r.SubmissionStatus,
		RejectionReason:  r.RejectionReason,
		MinimumAmount:    r.MinimumAmount,
		BillingDay:       r.BillingDay,
		CreatedAt:        r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SocialProgramSubscription) toSocialProgramSubscriptionResponse() SocialProgramSubscriptionResponse {
	return SocialProgramSubscriptionResponse{
		ID:              r.ID.String(),
		SocialProgramID: r.SocialProgramID.String(),
		AccountID:       r.AccountID.String(),
		Amount:          r.Amount,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toSocialProgramListResponse(programs []SocialProgram, pagination pkg.CursorPagination) SocialProgramListResponse {
	var responses []SocialProgramResponse
	for _, program := range programs {
		responses = append(responses, program.toSocialProgramResponse())
	}
	if responses == nil {
		responses = []SocialProgramResponse{}
	}
	return SocialProgramListResponse{
		SocialPrograms: responses,
		Pagination:     pagination,
	}
}