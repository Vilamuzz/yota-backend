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
	ProfilePicture string     `json:"profile_picture" gorm:"not null"`
	Gender         Gender     `json:"gender" gorm:"not null"`
	IsGraduated    bool       `json:"is_graduated" gorm:"not null"`
	Category       Category   `json:"category"`
	BirthDate      time.Time  `json:"birth_date"`
	BirthPlace     string     `json:"birth_place"`
	Address        string     `json:"address"`
	FamilyCard     string     `json:"family_card" gorm:"not null"`
	SKTM           string     `json:"sktm" gorm:"not null"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at" gorm:"index"`

	Achivements []Achivement `json:"achivements" gorm:"foreignKey:FosterChildrenID"`
}

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

type Achivement struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	FosterChildrenID uuid.UUID `json:"foster_children_id" gorm:"not null"`
	URL              string    `json:"url" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"not null"`
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
	ProfilePicture   string    `json:"profile_picture" gorm:"not null"`
	Gender           Gender    `json:"gender" gorm:"not null"`
	Category         Category  `json:"category"`
	BirthDate        time.Time `json:"birth_date"`
	BirthPlace       string    `json:"birth_place"`
	Address          string    `json:"address"`
	FamilyCard       string    `json:"family_card" gorm:"not null"`
	SKTM             string    `json:"sktm" gorm:"not null"`
	SubmitterName    string    `json:"submitter_name"`
	SubmitterPhone   string    `json:"submitter_phone"`
	SubmitterAddress string    `json:"submitter_address"`
	SubmitterIDCard  string    `json:"submitter_id_card"`
	SubmittedBy      uuid.UUID `json:"submitted_by" gorm:"not null"`
	Status           Status    `json:"status"`
	RejectionReason  string    `json:"rejection_reason"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Account account.Account `gorm:"foreignKey:SubmittedBy;references:ID"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
	StatusCanceled Status = "canceled"
)
