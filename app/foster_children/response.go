package foster_children

import "github.com/Vilamuzz/yota-backend/pkg"

type AchievementResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type FosterChildrenResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	ProfilePicture string                `json:"profile_picture"`
	Gender         Gender                `json:"gender"`
	IsGraduated    bool                  `json:"is_graduated"`
	Category       Category              `json:"category"`
	BirthDate      string                `json:"birth_date"`
	BirthPlace     string                `json:"birth_place"`
	Address        string                `json:"address"`
	FamilyCard     string                `json:"family_card"`
	SKTM           string                `json:"sktm"`
	Achievements   []AchievementResponse `json:"achievements"`
}

type FosterChildrenListResponse struct {
	FosterChildren []FosterChildrenResponse `json:"foster_children"`
	Pagination     pkg.CursorPagination     `json:"pagination"`
}

func (f *FosterChildren) ToFosterChildrenResponse() FosterChildrenResponse {
	var achievements []AchievementResponse
	for _, a := range f.Achivements {
		achievements = append(achievements, AchievementResponse{
			ID:  a.ID.String(),
			URL: a.URL,
		})
	}
	if achievements == nil {
		achievements = []AchievementResponse{}
	}

	return FosterChildrenResponse{
		ID:             f.ID.String(),
		Name:           f.Name,
		ProfilePicture: f.ProfilePicture,
		Gender:         f.Gender,
		IsGraduated:    f.IsGraduated,
		Category:       f.Category,
		BirthDate:      f.BirthDate.Format("2006-01-02"),
		BirthPlace:     f.BirthPlace,
		Address:        f.Address,
		FamilyCard:     f.FamilyCard,
		SKTM:           f.SKTM,
		Achievements:   achievements,
	}
}

func ToFosterChildrenListResponse(fosterChildren []FosterChildren, pagination pkg.CursorPagination) FosterChildrenListResponse {
	var responses []FosterChildrenResponse
	for _, f := range fosterChildren {
		responses = append(responses, f.ToFosterChildrenResponse())
	}
	if responses == nil {
		responses = []FosterChildrenResponse{}
	}
	return FosterChildrenListResponse{
		FosterChildren: responses,
		Pagination:     pagination,
	}
}

type FosterChildrenCandidateResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ProfilePicture   string `json:"profile_picture"`
	Gender           string `json:"gender"`
	Category         string `json:"category"`
	BirthDate        string `json:"birth_date"`
	BirthPlace       string `json:"birth_place"`
	Address          string `json:"address"`
	FamilyCard       string `json:"family_card"`
	SKTM             string `json:"sktm"`
	SubmitterName    string `json:"submitter_name"`
	SubmitterPhone   string `json:"submitter_phone"`
	SubmitterAddress string `json:"submitter_address"`
	SubmitterIDCard  string `json:"submitter_id_card"`
	SubmittedBy      string `json:"submitted_by"`
	Status           string `json:"status"`
	RejectionReason  string `json:"rejection_reason"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	AccountUsername  string `json:"account_username,omitempty"`
}

type FosterChildrenCandidateListResponse struct {
	Candidates []FosterChildrenCandidateResponse `json:"candidates"`
	Pagination pkg.CursorPagination              `json:"pagination"`
}

func (c *FosterChildrenCandidate) ToFosterChildrenCandidateResponse() FosterChildrenCandidateResponse {
	accountUsername := ""
	if c.Account.UserProfile.Username != "" {
		accountUsername = c.Account.UserProfile.Username
	}

	return FosterChildrenCandidateResponse{
		ID:               c.ID.String(),
		Name:             c.Name,
		ProfilePicture:   c.ProfilePicture,
		Gender:           string(c.Gender),
		Category:         string(c.Category),
		BirthDate:        c.BirthDate.Format("2006-01-02"),
		BirthPlace:       c.BirthPlace,
		Address:          c.Address,
		FamilyCard:       c.FamilyCard,
		SKTM:             c.SKTM,
		SubmitterName:    c.SubmitterName,
		SubmitterPhone:   c.SubmitterPhone,
		SubmitterAddress: c.SubmitterAddress,
		SubmitterIDCard:  c.SubmitterIDCard,
		SubmittedBy:      c.SubmittedBy.String(),
		Status:           string(c.Status),
		RejectionReason:  c.RejectionReason,
		CreatedAt:        c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        c.UpdatedAt.Format("2006-01-02 15:04:05"),
		AccountUsername:  accountUsername,
	}
}

func ToFosterChildrenCandidateListResponse(candidates []FosterChildrenCandidate, pagination pkg.CursorPagination) FosterChildrenCandidateListResponse {
	var responses []FosterChildrenCandidateResponse
	for _, c := range candidates {
		responses = append(responses, c.ToFosterChildrenCandidateResponse())
	}
	if responses == nil {
		responses = []FosterChildrenCandidateResponse{}
	}
	return FosterChildrenCandidateListResponse{
		Candidates: responses,
		Pagination: pagination,
	}
}
