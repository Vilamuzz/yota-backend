package foster_children_candidate

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type FosterChildrenCandidate struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"not null"`
	ProfilePicture   string    `json:"profilePicture" gorm:"not null"`
	Gender           Gender    `json:"gender" gorm:"not null"`
	Category         Category  `json:"category"`
	BirthDate        time.Time `json:"birthDate"`
	BirthPlace       string    `json:"birthPlace"`
	SchoolName       string    `json:"schoolName"`
	EducationLevel   int       `json:"educationLevel"`
	Address          string    `json:"address"`
	FamilyCard       string    `json:"familyCard" gorm:"not null"`
	SKTM             string    `json:"sktm" gorm:"not null"`
	SubmitterName    string    `json:"submitterName"`
	SubmitterPhone   string    `json:"submitterPhone"`
	SubmitterAddress string    `json:"submitterAddress"`
	SubmitterIDCard  string    `json:"submitterIdCard"`
	SubmittedBy      uuid.UUID `json:"submittedBy" gorm:"not null"`
	Status           Status    `json:"status"`
	RejectionReason  string    `json:"rejectionReason"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	Account account.Account `gorm:"foreignKey:SubmittedBy;references:ID"`
}

type Gender string

const (
	Male   Gender = "laki-laki"
	Female Gender = "perempuan"
)

type Category string

const (
	CategoryFatherless Category = "yatim"
	CategoryMotherless Category = "piatu"
	CategoryOrphan     Category = "yatim piatu"
)

type Status string

const (
	StatusPending               Status = "pending"
	StatusSocialManagerAccepted Status = "social_manager_accepted"
	StatusAccepted              Status = "accepted"
	StatusRejected              Status = "rejected"
	StatusCancelled             Status = "cancelled"
)
