package donation_program

import (
	"time"

	"github.com/google/uuid"
)

type DonationProgram struct {
	ID          uuid.UUID  `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"not null;index:idx_slug,type:btree"`
	CoverImage  string     `json:"coverImage"`
	Category    Category   `json:"category"`
	Description string     `json:"description"`
	FundTarget  float64    `json:"fundTarget"`
	Status      Status     `json:"status" gorm:"type:varchar(20);index:idx_status,type:btree;not null;default:'draft'"`
	StartDate   time.Time  `json:"startDate"`
	EndDate     time.Time  `json:"endDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt" gorm:"index"`

	CollectedFund float64 `json:"collectedFund" gorm:"->"`
	TotalExpense  float64 `json:"totalExpense" gorm:"->"`
}

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusExpired   Status = "expired"
	StatusArchived  Status = "archived"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusDraft, StatusCompleted, StatusExpired, StatusArchived:
		return true
	}
	return false
}

type Category string

const (
	CategoryEducation   Category = "pendidikan"
	CategoryHealth      Category = "kesehatan"
	CategoryEnvironment Category = "lingkungan"
	CategorySocial      Category = "sosial"
	CategoryDisaster    Category = "bencana"
	CategoryHumanity    Category = "kemanusiaan"
	CategoryOther       Category = "lainnya"
)

func (c Category) IsValid() bool {
	switch c {
	case CategoryEducation, CategoryHealth, CategoryEnvironment, CategorySocial, CategoryDisaster, CategoryHumanity, CategoryOther:
		return true
	}
	return false
}
