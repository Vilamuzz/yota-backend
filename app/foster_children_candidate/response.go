package foster_children_candidate

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type FosterChildrenCandidateResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ProfilePicture   string    `json:"profilePicture"`
	Gender           string    `json:"gender"`
	Category         string    `json:"category"`
	BirthDate        string    `json:"birthDate"`
	BirthPlace       string    `json:"birthPlace"`
	SchoolName       string    `json:"schoolName"`
	EducationLevel   int       `json:"educationLevel"`
	Address          string    `json:"address"`
	FamilyCard       string    `json:"familyCard"`
	SKTM             string    `json:"sktm"`
	SubmitterName    string    `json:"submitterName"`
	SubmitterPhone   string    `json:"submitterPhone"`
	SubmitterAddress string    `json:"submitterAddress"`
	SubmitterIDCard  string    `json:"submitterIdCard"`
	SubmittedBy      string    `json:"submittedBy"`
	Status           string    `json:"status"`
	RejectionReason  string    `json:"rejectionReason"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	AccountUsername  string    `json:"accountUsername,omitempty"`
}

type FosterChildrenCandidateListResponse struct {
	FosterChildrenCandidates []FosterChildrenCandidateResponse `json:"fosterChildrenCandidates"`
	Pagination               pkg.CursorPagination              `json:"pagination"`
}

type FosterChildrenCandidateAdminListResponse struct {
	FosterChildrenCandidates []FosterChildrenCandidateResponse `json:"fosterChildrenCandidates"`
	Pagination               pkg.OffsetPagination              `json:"pagination"`
}

func (c *FosterChildrenCandidate) ToFosterChildrenCandidateResponse() FosterChildrenCandidateResponse {
	accountUsername := ""
	if c.Account.UserProfile.Username != "" {
		accountUsername = c.Account.UserProfile.Username
	}

	return FosterChildrenCandidateResponse{
		ID:               c.ID.String(),
		Name:             c.Name,
		ProfilePicture:   s3_pkg.GetCDNURL(c.ProfilePicture),
		Gender:           string(c.Gender),
		Category:         string(c.Category),
		BirthDate:        c.BirthDate.Format("2006-01-02"),
		BirthPlace:       c.BirthPlace,
		Address:          c.Address,
		SchoolName:       c.SchoolName,
		EducationLevel:   c.EducationLevel,
		FamilyCard:       s3_pkg.GetCDNURL(c.FamilyCard),
		SKTM:             s3_pkg.GetCDNURL(c.SKTM),
		SubmitterName:    c.SubmitterName,
		SubmitterPhone:   c.SubmitterPhone,
		SubmitterAddress: c.SubmitterAddress,
		SubmitterIDCard:  s3_pkg.GetCDNURL(c.SubmitterIDCard),
		SubmittedBy:      c.SubmittedBy.String(),
		Status:           string(c.Status),
		RejectionReason:  c.RejectionReason,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
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

func ToFosterChildrenCandidateAdminListResponse(candidates []FosterChildrenCandidate, pagination pkg.OffsetPagination) FosterChildrenCandidateAdminListResponse {
	var responses []FosterChildrenCandidateResponse
	for _, c := range candidates {
		responses = append(responses, c.ToFosterChildrenCandidateResponse())
	}
	if responses == nil {
		responses = []FosterChildrenCandidateResponse{}
	}
	return FosterChildrenCandidateAdminListResponse{
		FosterChildrenCandidates: responses,
		Pagination:               pagination,
	}
}
