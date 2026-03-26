package social_program

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
	Status        Status  `json:"status"`
	IsSubscribed  bool    `json:"is_subscribed"`
	MinimumAmount float64 `json:"minimum_amount"`
	BillingDay    int     `json:"billing_day"`
	CreatedAt     string  `json:"created_at"`
}

type SocialProgramListResponse struct {
	Programs   []SocialProgramResponse `json:"programs"`
	Pagination pkg.CursorPagination    `json:"pagination"`
}

func (r *SocialProgram) toSocialProgramResponse() SocialProgramResponse {
	return SocialProgramResponse{
		ID:            r.ID,
		Title:         r.Title,
		Description:   r.Description,
		ImageURL:      r.ImageURL,
		Status:        r.Status,
		IsSubscribed:  r.IsSubscribed,
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
		Programs:   responses,
		Pagination: pagination,
	}
}
