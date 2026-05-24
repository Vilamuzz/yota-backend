package foster_children

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type FosterChildren struct {
	ID             uuid.UUID  `json:"id" gorm:"primaryKey"`
	Slug           string     `json:"slug" gorm:"not null"`
	Name           string     `json:"name" gorm:"not null"`
	ProfilePicture string     `json:"profilePicture" gorm:"not null"`
	Gender         Gender     `json:"gender" gorm:"not null"`
	IsGraduated    bool       `json:"isGraduated" gorm:"not null"`
	Category       Category   `json:"category"`
	BirthDate      time.Time  `json:"birthDate"`
	BirthPlace     string     `json:"birthPlace"`
	SchoolName     string     `json:"schoolName"`
	EducationLevel int        `json:"educationLevel"`
	Address        string     `json:"address"`
	FamilyCard     string     `json:"familyCard" gorm:"not null"`
	SKTM           string     `json:"sktm" gorm:"not null"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt" gorm:"index"`

	Achivements   []Achivement `json:"achivements" gorm:"foreignKey:FosterChildrenID"`
	CollectedFund float64      `json:"collectedFund" gorm:"->"`
	TotalExpense  float64      `json:"totalExpense" gorm:"->"`
}

type Gender string

const (
	Male   Gender = "laki-laki"
	Female Gender = "perempuan"
)

type Achivement struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	FosterChildrenID uuid.UUID `json:"fosterChildrenId" gorm:"not null"`
	URL              string    `json:"url" gorm:"not null"`
	Note             string    `json:"note"`
	CreatedAt        time.Time `json:"createdAt" gorm:"not null"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"not null"`
}

type Category string

const (
	CategoryFatherless Category = "yatim"
	CategoryMotherless Category = "piatu"
	CategoryOrphan     Category = "yatim piatu"
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

type Status string

const (
	StatusPending               Status = "pending"
	StatusSocialManagerAccepted Status = "social_manager_accepted"
	StatusAccepted              Status = "accepted"
	StatusRejected              Status = "rejected"
	StatusCancelled             Status = "cancelled"
)
