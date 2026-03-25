package log

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *Log) error
	FindAll(ctx context.Context, options map[string]interface{}) ([]Log, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Log, error) {
	query := r.db.WithContext(ctx).Model(&Log{}).Order("created_at DESC")

	if v, ok := options["entity_type"].(string); ok && v != "" {
		query = query.Where("entity_type = ?", v)
	}
	if v, ok := options["entity_id"].(string); ok && v != "" {
		query = query.Where("entity_id = ?", v)
	}
	if v, ok := options["user_id"].(string); ok && v != "" {
		query = query.Where("user_id = ?", v)
	}
	if v, ok := options["action"].(string); ok && v != "" {
		query = query.Where("action = ?", v)
	}

	limit, _ := options["limit"].(int)
	if limit <= 0 {
		limit = 20
	}
	query = query.Limit(limit + 1)

	if cursor, ok := options["next_cursor"].(string); ok && cursor != "" {
		data, err := pkg.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", data.CreatedAt, data.ID)
		}
	}

	var logs []Log
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
