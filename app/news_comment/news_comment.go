package news_comment

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/google/uuid"
)

type NewsComment struct {
	ID              uuid.UUID  `json:"id" gorm:"primaryKey"`
	NewsID          uuid.UUID  `json:"news_id" gorm:"index;not null"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id"`
	AccountID       uuid.UUID  `json:"account_id" gorm:"index;not null"`
	Content         string     `json:"content" gorm:"not null"`
	CreatedAt       time.Time  `json:"created_at"`
	DeletedAt       *time.Time `json:"deleted_at" gorm:"index"`

	News               news.News           `gorm:"foreignKey:NewsID"`
	ParentComment      *NewsComment        `gorm:"foreignKey:ParentCommentID"`
	NewsCommentReports []NewsCommentReport `gorm:"foreignKey:NewsCommentID"`
}

type NewsCommentReport struct {
	AccountID     uuid.UUID `json:"account_id" gorm:"primaryKey"`
	NewsCommentID uuid.UUID `json:"news_comment_id" gorm:"primaryKey"`
	Reason        string    `json:"reason" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
}
