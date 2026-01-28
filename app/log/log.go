package log

import "time"

type Log struct {
	ID         string    `json:"id" gorm:"primary_key"`
	UserID     string    `json:"user_id" gorm:"not null"`
	Action     string    `json:"action" gorm:"not null"`
	EntityType string    `json:"entity_type" gorm:"not null"`
	EntityID   string    `json:"entity_id" gorm:"not null"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	CreatedAt  time.Time `json:"created_at"`
}
