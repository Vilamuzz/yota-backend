package log

import "time"

type Log struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     *string   `json:"userId" gorm:"index"`
	Action     string    `json:"action" gorm:"not null"`
	EntityType string    `json:"entityType" gorm:"not null"`
	EntityID   string    `json:"entityId" gorm:"not null"`
	OldValue   string    `json:"oldValue"`
	NewValue   string    `json:"newValue"`
	CreatedAt  time.Time `json:"createdAt"`
}
