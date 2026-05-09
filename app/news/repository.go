package news

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllNews(ctx context.Context, options map[string]interface{}) ([]News, error)
	FindOneNews(ctx context.Context, options map[string]interface{}) (*News, error)
	CreateNews(ctx context.Context, news *News) error
	UpdateNews(ctx context.Context, id string, updateData map[string]interface{}) error
	UpdateNewsMedia(ctx context.Context, news *News, media []media.Media) error
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

func (r *repository) FindAllNews(ctx context.Context, options map[string]interface{}) ([]News, error) {
	var newsList []News
	query := r.Conn.WithContext(ctx).Preload("Media").Where("deleted_at IS NULL")

	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&newsList).Error; err != nil {
		return nil, err
	}
	return newsList, nil
}

func (r *repository) FindOneNews(ctx context.Context, options map[string]interface{}) (*News, error) {
	var news News
	query := r.Conn.WithContext(ctx).Preload("Media").Where("deleted_at IS NULL")
	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("slug = ?", slug)
	}
	if published, ok := options["published"]; ok && published.(bool) {
		query = query.Where("status = ?", media.MediaStatusPublished)
	}
	if err := query.First(&news).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *repository) CreateNews(ctx context.Context, news *News) error {
	return r.Conn.WithContext(ctx).Create(news).Error
}

func (r *repository) UpdateNews(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) UpdateNewsMedia(ctx context.Context, news *News, media []media.Media) error {
	return r.Conn.WithContext(ctx).Model(news).Association("Media").Replace(media)
}

func (r *repository) DeleteNews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *repository) IncrementViews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
