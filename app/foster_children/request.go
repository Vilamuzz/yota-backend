package foster_children

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateFosterChildrenRequest struct {
	Name           string                  `form:"name"`
	Gender         Gender                  `form:"gender"`
	IsGraduated    bool                    `form:"is_graduated"`
	Category       Category                `form:"category"`
	BirthDate      string                  `form:"birth_date"`
	BirthPlace     string                  `form:"birth_place"`
	Address        string                  `form:"address"`
	ProfilePicture *multipart.FileHeader   `form:"profile_picture" swaggerignore:"true"`
	FamilyCard     *multipart.FileHeader   `form:"family_card" swaggerignore:"true"`
	SKTM           *multipart.FileHeader   `form:"sktm" swaggerignore:"true"`
	Achievements   []*multipart.FileHeader `form:"achievements" swaggerignore:"true"`
}

type UpdateFosterChildrenRequest struct {
	Name           string                  `form:"name"`
	Gender         Gender                  `form:"gender"`
	IsGraduated    *bool                   `form:"is_graduated"`
	Category       Category                `form:"category"`
	BirthDate      string                  `form:"birth_date"`
	BirthPlace     string                  `form:"birth_place"`
	Address        string                  `form:"address"`
	ProfilePicture *multipart.FileHeader   `form:"profile_picture" swaggerignore:"true"`
	FamilyCard     *multipart.FileHeader   `form:"family_card" swaggerignore:"true"`
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
	BirthDate        string                `form:"birth_date"`
	BirthPlace       string                `form:"birth_place"`
	Address          string                `form:"address"`
	ProfilePicture   *multipart.FileHeader `form:"profile_picture" swaggerignore:"true"`
	FamilyCard       *multipart.FileHeader `form:"family_card" swaggerignore:"true"`
	SKTM             *multipart.FileHeader `form:"sktm" swaggerignore:"true"`
	SubmitterName    string                `form:"submitter_name"`
	SubmitterPhone   string                `form:"submitter_phone"`
	SubmitterAddress string                `form:"submitter_address"`
	SubmitterIDCard  *multipart.FileHeader `form:"submitter_id_card" swaggerignore:"true"`
}

type UpdateFosterChildrenCandidateStatusRequest struct {
	Status          Status `json:"status"`
	RejectionReason string `json:"rejection_reason"`
}

type FosterChildrenCandidateQueryParams struct {
	Status    Status `form:"status"`
	AccountID string `form:"account_id"`
	pkg.PaginationParams
}
