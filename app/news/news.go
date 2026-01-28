package news

import "time"

type News struct {
	ID          string    `json:"id" gorm:"primary_key"`
	Title       string    `json:"title" gorm:"not null"`
	Category    string    `json:"category" gorm:"not null"`
	Content     string    `json:"content" gorm:"not null"`
	Status      string    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Views       int       `json:"views" gorm:"not null;default:0"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}
