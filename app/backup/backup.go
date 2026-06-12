package backup

import (
	"time"

	"github.com/google/uuid"
)

type Backup struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Filename  string     `gorm:"type:varchar(255);unique;not null" json:"filename"`
	Size      int64      `gorm:"not null" json:"size"`
	Duration  int64      `gorm:"not null" json:"duration"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt"`
}
