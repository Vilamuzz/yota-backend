package foster_children

import (
	"time"

	"github.com/google/uuid"
)

type FosterChildren struct {
	ID             uuid.UUID  `json:"id" gorm:"primaryKey"`
	Slug           string     `json:"slug" gorm:"not null"`
	Name           string     `json:"name" gorm:"not null"`
	Nik            string     `json:"nik"`
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
