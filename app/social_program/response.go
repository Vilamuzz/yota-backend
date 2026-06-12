package social_program

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramDetailResponse struct {
	ID               string  `json:"id"`
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	CoverImage       string  `json:"coverImage"`
	Status           Status  `json:"status"`
	IsSubscribed     bool    `json:"isSubscribed"`
	SubscriptionId   string  `json:"subscriptionId"`
	TotalSubscribers int64   `json:"totalSubscribers"`
	MinimumAmount    float64 `json:"minimumAmount"`
	BillingDay       int     `json:"billingDay"`
	CreatedAt        string  `json:"createdAt"`
	TotalExpense     float64 `json:"totalExpense"`
}

type SocialProgramListItemResponse struct {
	ID               string  `json:"id"`
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	CoverImage       string  `json:"coverImage"`
	Status           Status  `json:"status"`
	IsSubscribed     bool    `json:"isSubscribed"`
	SubscriptionId   string  `json:"subscriptionId"`
	TotalSubscribers int64   `json:"totalSubscribers"`
	MinimumAmount    float64 `json:"minimumAmount"`
	BillingDay       int     `json:"billingDay"`
	TotalExpense     float64 `json:"totalExpense"`
}

type SocialProgramListResponse struct {
	SocialPrograms []SocialProgramListItemResponse `json:"socialPrograms"`
	Pagination     pkg.CursorPagination            `json:"pagination"`
}

func (r *SocialProgram) ToSocialProgramDetailResponse() SocialProgramDetailResponse {
	return SocialProgramDetailResponse{
		ID:               r.ID.String(),
		Slug:             r.Slug,
		Title:            r.Title,
		Description:      r.Description,
		CoverImage:       r.CoverImage,
		Status:           r.Status,
		TotalSubscribers: r.TotalSubscribers,
		IsSubscribed:     r.IsSubscribed,
		SubscriptionId:   r.SubscriptionID,
		MinimumAmount:    r.MinimumAmount,
		BillingDay:       r.BillingDay,
		CreatedAt:        r.CreatedAt.Format("2006-01-02 15:04:05"),
		TotalExpense:     r.TotalExpense,
	}
}

func (r *SocialProgram) ToSocialProgramListItemResponse() SocialProgramListItemResponse {
	return SocialProgramListItemResponse{
		ID:               r.ID.String(),
		Slug:             r.Slug,
		Title:            r.Title,
		CoverImage:       r.CoverImage,
		Status:           r.Status,
		TotalSubscribers: r.TotalSubscribers,
		IsSubscribed:     r.IsSubscribed,
		SubscriptionId:   r.SubscriptionID,
		MinimumAmount:    r.MinimumAmount,
		BillingDay:       r.BillingDay,
		TotalExpense:     r.TotalExpense,
	}
}

func ToSocialProgramListResponse(programs []SocialProgram, pagination pkg.CursorPagination) SocialProgramListResponse {
	var responses []SocialProgramListItemResponse
	for _, program := range programs {
		responses = append(responses, program.ToSocialProgramListItemResponse())
	}
	if responses == nil {
		responses = []SocialProgramListItemResponse{}
	}
	return SocialProgramListResponse{
		SocialPrograms: responses,
		Pagination:     pagination,
	}
}

type AdminSocialProgramDetailResponse struct {
	SocialProgramDetailResponse
	CollectedFund float64 `json:"collectedFund"`
}

type AdminSocialProgramListItemResponse struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Status           Status  `json:"status"`
	TotalSubscribers int64   `json:"totalSubscribers"`
	MinimumAmount    float64 `json:"minimumAmount"`
	CollectedFund    float64 `json:"collectedFund"`
	TotalExpense     float64 `json:"totalExpense"`
	CreatedAt        string  `json:"createdAt"`
}

type AdminSocialProgramListResponse struct {
	AdminSocialPrograms []AdminSocialProgramListItemResponse `json:"socialPrograms"`
	Pagination          pkg.CursorPagination                 `json:"pagination"`
}

func (a *SocialProgram) ToAdminSocialProgramDetailResponse() AdminSocialProgramDetailResponse {
	return AdminSocialProgramDetailResponse{
		SocialProgramDetailResponse: a.ToSocialProgramDetailResponse(),
		CollectedFund:               a.CollectedFund,
	}
}

func (a *SocialProgram) ToAdminSocialProgramListItemResponse() AdminSocialProgramListItemResponse {
	return AdminSocialProgramListItemResponse{
		ID:               a.ID.String(),
		Title:            a.Title,
		Status:           a.Status,
		TotalSubscribers: a.TotalSubscribers,
		MinimumAmount:    a.MinimumAmount,
		CollectedFund:    a.CollectedFund,
		TotalExpense:     a.TotalExpense,
		CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToAdminSocialProgramListResponse(programs []SocialProgram, pagination pkg.CursorPagination) AdminSocialProgramListResponse {
	var responses []AdminSocialProgramListItemResponse
	for _, a := range programs {
		responses = append(responses, a.ToAdminSocialProgramListItemResponse())
	}
	if responses == nil {
		responses = []AdminSocialProgramListItemResponse{}
	}
	return AdminSocialProgramListResponse{
		AdminSocialPrograms: responses,
		Pagination:          pagination,
	}
}
