package foster_children_candidate

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateFosterChildrenCandidateRequest struct {
	Name             string                `form:"name"`
	Gender           Gender                `form:"gender"`
	Category         Category              `form:"category"`
	BirthDate        string                `form:"birthDate"`
	BirthPlace       string                `form:"birthPlace"`
	SchoolName       string                `form:"schoolName"`
	EducationLevel   int                   `form:"educationLevel"`
	Address          string                `form:"address"`
	ProfilePicture   *multipart.FileHeader `form:"profilePicture" swaggerignore:"true"`
	FamilyCard       *multipart.FileHeader `form:"familyCard" swaggerignore:"true"`
	SKTM             *multipart.FileHeader `form:"sktm" swaggerignore:"true"`
	SubmitterName    string                `form:"submitterName"`
	SubmitterPhone   string                `form:"submitterPhone"`
	SubmitterAddress string                `form:"submitterAddress"`
	SubmitterIDCard  *multipart.FileHeader `form:"submitterIdCard" swaggerignore:"true"`
}

type RejectFosterChildrenCandidateRequest struct {
	RejectionReason string `json:"rejectionReason"`
}

type FosterChildrenCandidateQueryParams struct {
	Search    string `form:"search"`
	SortBy    string `form:"sortBy"`
	Status    Status `form:"status"`
	AccountID string `form:"accountId"`
	pkg.PaginationParams
}

type FosterChildrenCandidateAdminQueryParams struct {
	Status    Status   `form:"status"`
	AccountID string   `form:"accountId"`
	Category  Category `form:"category"`
	Gender    Gender   `form:"gender"`
	SortBy    string   `form:"sortBy"`
	Search    string   `form:"search"`
	Page      int      `form:"page"`
	Limit     int      `form:"limit"`
}
