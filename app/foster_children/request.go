package foster_children

import (
	"mime/multipart"
)

type CreateFosterChildrenRequest struct {
	Name            string                  `form:"name"`
	Gender          Gender                  `form:"gender"`
	IsGraduated     bool                    `form:"isGraduated"`
	Category        Category                `form:"category"`
	BirthDate       string                  `form:"birthDate"`
	BirthPlace      string                  `form:"birthPlace"`
	SchoolName      string                  `form:"schoolName"`
	EducationLevel  int                     `form:"educationLevel"`
	Address         string                  `form:"address"`
	ProfilePicture  *multipart.FileHeader   `form:"profilePicture" swaggerignore:"true"`
	FamilyCard      *multipart.FileHeader   `form:"familyCard" swaggerignore:"true"`
	SKTM            *multipart.FileHeader   `form:"sktm" swaggerignore:"true"`
	Achievements    []*multipart.FileHeader `form:"achievements[]" swaggerignore:"true"`
	AchivementNotes []string                `form:"achivementNotes[]"`
}

type UpdateFosterChildrenRequest struct {
	Name                  string                  `form:"name"`
	Gender                Gender                  `form:"gender"`
	IsGraduated           *bool                   `form:"isGraduated"`
	Category              Category                `form:"category"`
	BirthDate             string                  `form:"birthDate"`
	BirthPlace            string                  `form:"birthPlace"`
	SchoolName            string                  `form:"schoolName"`
	EducationLevel        int                     `form:"educationLevel"`
	Address               string                  `form:"address"`
	ProfilePicture        *multipart.FileHeader   `form:"profilePicture" swaggerignore:"true"`
	FamilyCard            *multipart.FileHeader   `form:"familyCard" swaggerignore:"true"`
	SKTM                  *multipart.FileHeader   `form:"sktm" swaggerignore:"true"`
	Achievements          []*multipart.FileHeader `form:"achievements[]" swaggerignore:"true"`
	AchivementNotes       []string                `form:"achivementNotes[]"`
	AchivementIDs         []string                `form:"achivementIDs[]"`
	UpdateAchivementNotes []string                `form:"updateAchivementNotes[]"`
}

type FosterChildrenQueryParams struct {
	Search         string   `form:"search"`
	Category       Category `form:"category"`
	Gender         Gender   `form:"gender"`
	IsGraduated    *bool    `form:"isGraduated"`
	SortBy         string   `form:"sortBy"` // e.g. "name asc", "education_level desc", "created_at desc"
	Page           int      `form:"page"`
	Limit          int      `form:"limit"`
}

