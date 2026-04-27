package donation_program

import (
	"time"

	"github.com/google/uuid"
)

type DonationProgram struct {
	ID          uuid.UUID  `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"not null;unique"`
	CoverImage  string     `json:"coverImage"`
	Category    Category   `json:"category" gorm:"not null"`
	Description string     `json:"description" gorm:"not null"`
	FundTarget  float64    `json:"fundTarget" gorm:"not null"`
	Status      Status     `json:"status" gorm:"type:varchar(20);index;not null;default:'active'"`
	StartDate   time.Time  `json:"startDate" gorm:"not null"`
	EndDate     time.Time  `json:"endDate" gorm:"not null"`
	PublishedAt *time.Time `json:"publishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt" gorm:"index"`

	CollectedFund float64 `json:"collectedFund" gorm:"-"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusDraft     Status = "draft"
	StatusCompleted Status = "complete"
	StatusExpired   Status = "expired"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusDraft, StatusCompleted, StatusExpired:
		return true
	}
	return false
}

type Category string

const (
	CategoryEducation   Category = "education"
	CategoryHealth      Category = "health"
	CategoryEnvironment Category = "environment"
	CategorySocial      Category = "social"
	CategoryDisaster    Category = "disaster"
	CategoryHumanity    Category = "humanity"
	CategoryOther       Category = "other"
)

func (c Category) IsValid() bool {
	switch c {
	case CategoryEducation, CategoryHealth, CategoryEnvironment, CategorySocial, CategoryDisaster, CategoryHumanity, CategoryOther:
		return true
	}
	return false
}
