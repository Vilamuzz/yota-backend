package social_program

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramResponse struct {
	ID            string  `json:"id"`
	Slug          string  `json:"slug"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	CoverImage    string  `json:"coverImage"`
	Status        Status  `json:"status"`
	IsSubscribed  bool    `json:"isSubscribed"`
	MinimumAmount float64 `json:"minimumAmount"`
	BillingDay    int     `json:"billingDay"`
	CreatedAt     string  `json:"createdAt"`
}

type SocialProgramListResponse struct {
	SocialPrograms []SocialProgramResponse `json:"socialPrograms"`
	Pagination     pkg.CursorPagination    `json:"pagination"`
}

func (r *SocialProgram) toSocialProgramResponse() SocialProgramResponse {
	return SocialProgramResponse{
		ID:            r.ID.String(),
		Slug:          r.Slug,
		Title:         r.Title,
		Description:   r.Description,
		CoverImage:    r.CoverImage,
		Status:        r.Status,
		MinimumAmount: r.MinimumAmount,
		BillingDay:    r.BillingDay,
		CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
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
