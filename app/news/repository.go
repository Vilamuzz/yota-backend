package news

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FetchAllNews(ctx context.Context, options map[string]interface{}) ([]News, error)
	FetchNewsByID(ctx context.Context, id string) (*News, error)
	CreateOneNews(ctx context.Context, news *News) error
	UpdateNews(ctx context.Context, id string, updateData map[string]interface{}) error
	DeleteNews(ctx context.Context, id string) error
	IncrementViews(ctx context.Context, id string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FetchAllNews(ctx context.Context, options map[string]interface{}) ([]News, error) {
	var newsList []News
	query := r.Conn.WithContext(ctx)

	// Apply filters
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Apply cursor-based pagination
	if cursor, ok := options["cursor"]; ok && cursor != "" {
		cursorStr := cursor.(string)
		cursorData, err := pkg.DecodeCursor(cursorStr)
		if err == nil {
			// Cursor format: created_at|id
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	// Apply limit (fetch one extra to check if there's a next page)
	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	query = query.Limit(limit + 1)

	// Order by created date (newest first) and ID for consistent ordering
	query = query.Order("created_at DESC, id DESC")

	if err := query.Find(&newsList).Error; err != nil {
		return nil, err
	}

	return newsList, nil
}

func (r *repository) FetchNewsByID(ctx context.Context, id string) (*News, error) {
	var news News
	if err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&news).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *repository) CreateOneNews(ctx context.Context, news *News) error {
	return r.Conn.WithContext(ctx).Create(news).Error
}

func (r *repository) UpdateNews(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) DeleteNews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&News{}).Error
}

func (r *repository) IncrementViews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
