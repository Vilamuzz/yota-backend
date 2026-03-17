package donation

import "time"

type Donation struct {
	ID            string     `json:"id" gorm:"primary_key"`
	Title         string     `json:"title" gorm:"not null"`
	Slug          string     `json:"slug" gorm:"not null;unique"`
	Description   string     `json:"description" gorm:"not null"`
	ImageURL      string     `json:"image_url"`
	Category      Category   `json:"category" gorm:"not null"`
	FundTarget    float64    `json:"fund_target" gorm:"not null"`
	CollectedFund float64    `json:"collected_fund" gorm:"->;column:collected_fund"`
	Status        Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	DateEnd       time.Time  `json:"date_end"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     time.Time  `json:"deleted_at"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusDraft     Status = "draft"
	StatusCompleted Status = "complete"
	StatusExpired   Status = "expired"
)

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
