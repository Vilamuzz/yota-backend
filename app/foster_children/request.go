package foster_children

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateFosterChildrenRequest struct {
	Name           string                  `form:"name"`
	Gender         Gender                  `form:"gender"`
	IsGraduated    bool                    `form:"isGraduated"`
	Category       Category                `form:"category"`
	BirthDate      string                  `form:"birthDate"`
	BirthPlace     string                  `form:"birthPlace"`
	Address        string                  `form:"address"`
	ProfilePicture *multipart.FileHeader   `form:"profilePicture" swaggerignore:"true"`
	FamilyCard     *multipart.FileHeader   `form:"familyCard" swaggerignore:"true"`
	SKTM           *multipart.FileHeader   `form:"sktm" swaggerignore:"true"`
	Achievements   []*multipart.FileHeader `form:"achievements" swaggerignore:"true"`
}

type UpdateFosterChildrenRequest struct {
	Name           string                  `form:"name"`
	Gender         Gender                  `form:"gender"`
	IsGraduated    *bool                   `form:"isGraduated"`
	Category       Category                `form:"category"`
	BirthDate      string                  `form:"birthDate"`
	BirthPlace     string                  `form:"birthPlace"`
	Address        string                  `form:"address"`
	ProfilePicture *multipart.FileHeader   `form:"profilePicture" swaggerignore:"true"`
	FamilyCard     *multipart.FileHeader   `form:"familyCard" swaggerignore:"true"`
	SKTM           *multipart.FileHeader   `form:"sktm" swaggerignore:"true"`
	Achievements   []*multipart.FileHeader `form:"achievements" swaggerignore:"true"`
}

type FosterChildrenQueryParams struct {
	Search   string `form:"search"`
	Category string `form:"category"`
	pkg.PaginationParams
}

type CreateFosterChildrenCandidateRequest struct {
	Name             string                `form:"name"`
	Gender           Gender                `form:"gender"`
	Category         Category              `form:"category"`
	BirthDate        string                `form:"birthDate"`
	BirthPlace       string                `form:"birthPlace"`
	Address          string                `form:"address"`
	ProfilePicture   *multipart.FileHeader `form:"profilePicture" swaggerignore:"true"`
	FamilyCard       *multipart.FileHeader `form:"familyCard" swaggerignore:"true"`
	SKTM             *multipart.FileHeader `form:"sktm" swaggerignore:"true"`
	SubmitterName    string                `form:"submitterName"`
	SubmitterPhone   string                `form:"submitterPhone"`
	SubmitterAddress string                `form:"submitterAddress"`
	SubmitterIDCard  *multipart.FileHeader `form:"submitterIdCard" swaggerignore:"true"`
}

type UpdateFosterChildrenCandidateStatusRequest struct {
	Status          Status `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}

type FosterChildrenCandidateQueryParams struct {
	Status    Status `form:"status"`
	AccountID string `form:"accountId"`
	pkg.PaginationParams
}
