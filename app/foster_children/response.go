package foster_children

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type AchievementResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type FosterChildrenDetailResponse struct {
	ID             string                `json:"id"`
	Slug           string                `json:"slug"`
	Name           string                `json:"name"`
	ProfilePicture string                `json:"profilePicture"`
	Gender         Gender                `json:"gender"`
	IsGraduated    bool                  `json:"isGraduated"`
	Category       Category              `json:"category"`
	BirthDate      string                `json:"birthDate"`
	BirthPlace     string                `json:"birthPlace"`
	SchoolName     string                `json:"schoolName"`
	EducationLevel int                   `json:"educationLevel"`
	Address        string                `json:"address"`
	Achievements   []AchievementResponse `json:"achievements"`
	CreatedAt      time.Time             `json:"createdAt"`
	TotalExpense   float64               `json:"totalExpense"`
}

type FosterChildrenListItemResponse struct {
	ID             string   `json:"id"`
	Slug           string   `json:"slug"`
	Name           string   `json:"name"`
	ProfilePicture string   `json:"profilePicture"`
	BirthDate      string   `json:"birthDate"`
	Gender         Gender   `json:"gender"`
	IsGraduated    bool     `json:"isGraduated"`
	Category       Category `json:"category"`
	TotalExpense   float64  `json:"totalExpense"`
}

type FosterChildrenListResponse struct {
	FosterChildren []FosterChildrenListItemResponse `json:"fosterChildren"`
	Pagination     pkg.OffsetPagination             `json:"pagination"`
}

func (f *FosterChildren) ToFosterChildrenDetailResponse() FosterChildrenDetailResponse {
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

	return FosterChildrenDetailResponse{
		ID:             f.ID.String(),
		Slug:           f.Slug,
		Name:           f.Name,
		ProfilePicture: f.ProfilePicture,
		Gender:         f.Gender,
		IsGraduated:    f.IsGraduated,
		Category:       f.Category,
		BirthDate:      f.BirthDate.Format("2006-01-02"),
		BirthPlace:     f.BirthPlace,
		SchoolName:     f.SchoolName,
		EducationLevel: f.EducationLevel,
		Address:        f.Address,
		Achievements:   achievements,
		CreatedAt:      f.CreatedAt,
		TotalExpense:   f.TotalExpense,
	}
}

func (f *FosterChildren) ToFosterChildrenListItemResponse() FosterChildrenListItemResponse {
	return FosterChildrenListItemResponse{
		ID:             f.ID.String(),
		Slug:           f.Slug,
		Name:           f.Name,
		ProfilePicture: f.ProfilePicture,
		BirthDate:      f.BirthDate.Format("2006-01-02"),
		Gender:         f.Gender,
		IsGraduated:    f.IsGraduated,
		Category:       f.Category,
		TotalExpense:   f.TotalExpense,
	}
}

func ToFosterChildrenListResponse(fosterChildren []FosterChildren, pagination pkg.OffsetPagination) FosterChildrenListResponse {
	var responses []FosterChildrenListItemResponse
	for _, f := range fosterChildren {
		responses = append(responses, f.ToFosterChildrenListItemResponse())
	}
	if responses == nil {
		responses = []FosterChildrenListItemResponse{}
	}
	return FosterChildrenListResponse{
		FosterChildren: responses,
		Pagination:     pagination,
	}
}

type AdminFosterChildrenDetailResponse struct {
	FosterChildrenDetailResponse
	FamilyCard    string  `json:"familyCard"`
	SKTM          string  `json:"sktm"`
	CollectedFund float64 `json:"collectedFund"`
}

type AdminFosterChildrenListItemResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profilePicture"`
	Gender         Gender    `json:"gender"`
	IsGraduated    bool      `json:"isGraduated"`
	Category       Category  `json:"category"`
	CollectedFund  float64   `json:"collectedFund"`
	TotalExpense   float64   `json:"totalExpense"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AdminFosterChildrenListResponse struct {
	AdminFosterChildren []AdminFosterChildrenListItemResponse `json:"fosterChildren"`
	Pagination          pkg.OffsetPagination                  `json:"pagination"`
}

func (a *FosterChildren) ToAdminFosterChildrenDetailResponse() AdminFosterChildrenDetailResponse {
	return AdminFosterChildrenDetailResponse{
		FosterChildrenDetailResponse: a.ToFosterChildrenDetailResponse(),
		FamilyCard:                   a.FamilyCard,
		SKTM:                         a.SKTM,
		CollectedFund:                a.CollectedFund,
	}
}

func (a *FosterChildren) ToAdminFosterChildrenListItemResponse() AdminFosterChildrenListItemResponse {
	return AdminFosterChildrenListItemResponse{
		ID:             a.ID.String(),
		Name:           a.Name,
		ProfilePicture: a.ProfilePicture,
		Gender:         a.Gender,
		IsGraduated:    a.IsGraduated,
		Category:       a.Category,
		CollectedFund:  a.CollectedFund,
		TotalExpense:   a.TotalExpense,
		CreatedAt:      a.CreatedAt,
	}
}

func ToAdminFosterChildrenListResponse(fosterChildren []FosterChildren, pagination pkg.OffsetPagination) AdminFosterChildrenListResponse {
	var responses []AdminFosterChildrenListItemResponse
	for _, a := range fosterChildren {
		responses = append(responses, a.ToAdminFosterChildrenListItemResponse())
	}
	if responses == nil {
		responses = []AdminFosterChildrenListItemResponse{}
	}
	return AdminFosterChildrenListResponse{
		AdminFosterChildren: responses,
		Pagination:          pagination,
	}
}

