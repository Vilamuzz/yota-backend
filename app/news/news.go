package news

import (
	"time"
)

type News struct {
	ID          string     `json:"id" gorm:"primary_key"`
	Title       string     `json:"title" gorm:"not null"`
	Category    Category   `json:"category" gorm:"not null"`
	Content     string     `json:"content" gorm:"not null"`
	Image       string     `json:"image"`
	Status      Status     `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	Views       int        `json:"views" gorm:"not null;default:0"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Category string

const (
	CategoryGeneral      Category = "general"
	CategoryEvent        Category = "event"
	CategoryAnnouncement Category = "announcement"
	CategoryDonation     Category = "donation"
	CategorySocial       Category = "social"
)
