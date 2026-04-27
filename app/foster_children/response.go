package foster_children

import "github.com/Vilamuzz/yota-backend/pkg"

type AchievementResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type FosterChildrenResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	ProfilePicture string                `json:"profilePicture"`
	Gender         Gender                `json:"gender"`
	IsGraduated    bool                  `json:"isGraduated"`
	Category       Category              `json:"category"`
	BirthDate      string                `json:"birthDate"`
	BirthPlace     string                `json:"birthPlace"`
	Address        string                `json:"address"`
	FamilyCard     string                `json:"familyCard"`
	SKTM           string                `json:"sktm"`
	Achievements   []AchievementResponse `json:"achievements"`
}

type FosterChildrenListResponse struct {
	FosterChildren []FosterChildrenResponse `json:"fosterChildren"`
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
	ProfilePicture   string `json:"profilePicture"`
	Gender           string `json:"gender"`
	Category         string `json:"category"`
	BirthDate        string `json:"birthDate"`
	BirthPlace       string `json:"birthPlace"`
	Address          string `json:"address"`
	FamilyCard       string `json:"familyCard"`
	SKTM             string `json:"sktm"`
	SubmitterName    string `json:"submitterName"`
	SubmitterPhone   string `json:"submitterPhone"`
	SubmitterAddress string `json:"submitterAddress"`
	SubmitterIDCard  string `json:"submitterIdCard"`
	SubmittedBy      string `json:"submittedBy"`
	Status           string `json:"status"`
	RejectionReason  string `json:"rejectionReason"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	AccountUsername  string `json:"accountUsername,omitempty"`
}

type FosterChildrenCandidateListResponse struct {
	FosterChildrenCandidates []FosterChildrenCandidateResponse `json:"fosterChildrenCandidates"`
	Pagination               pkg.CursorPagination              `json:"pagination"`
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
		FosterChildrenCandidates: responses,
		Pagination:               pagination,
	}
}
