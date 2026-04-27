package news_comment

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/google/uuid"
)

type NewsComment struct {
	ID              uuid.UUID  `json:"id" gorm:"primaryKey"`
	NewsID          uuid.UUID  `json:"newsId" gorm:"index;not null"`
	ParentCommentID *uuid.UUID `json:"parentCommentId"`
	AccountID       uuid.UUID  `json:"accountId" gorm:"index;not null"`
	Content         string     `json:"content" gorm:"not null"`
	CreatedAt       time.Time  `json:"createdAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`

	News               news.News           `gorm:"foreignKey:NewsID"`
	ParentComment      *NewsComment        `gorm:"foreignKey:ParentCommentID"`
	NewsCommentReports []NewsCommentReport `gorm:"foreignKey:NewsCommentID"`
}

type NewsCommentReport struct {
	AccountID     uuid.UUID `json:"accountId" gorm:"primaryKey"`
	NewsCommentID uuid.UUID `json:"newsCommentId" gorm:"primaryKey"`
	Reason        string    `json:"reason" gorm:"not null"`
	CreatedAt     time.Time `json:"createdAt"`
}
