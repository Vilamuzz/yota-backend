package gallery

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error)
	FindOneGallery(ctx context.Context, options map[string]interface{}) (*Gallery, error)
	CreateGallery(ctx context.Context, gallery *Gallery) error
	UpdateGallery(ctx context.Context, id string, updateData map[string]interface{}) error
	DeleteGallery(ctx context.Context, id string) error
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

func (r *repository) FindAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error) {
	var galleries []Gallery
	query := r.Conn.WithContext(ctx).Preload("Media").Where("deleted_at IS NULL")

	if categoryID, ok := options["category_id"]; ok && categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
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

	if _, isPrev := options["prev_cursor"]; isPrev {
		query = query.Order("created_at ASC, id ASC")
	} else {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&galleries).Error; err != nil {
		return nil, err
	}
	return galleries, nil
}

func (r *repository) FindOneGallery(ctx context.Context, options map[string]interface{}) (*Gallery, error) {
	var gallery Gallery
	query := r.Conn.WithContext(ctx).Preload("Media").Where("deleted_at IS NULL")
	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("slug = ?", slug)
	}
	if title, ok := options["title"]; ok && title != "" {
		query = query.Where("title = ?", title)
	}
	if published, ok := options["published"]; ok && published.(bool) {
		query = query.Where("status = ?", media.MediaStatusPublished)
	}
	if err := query.First(&gallery).Error; err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (r *repository) CreateGallery(ctx context.Context, gallery *Gallery) error {
	return r.Conn.WithContext(ctx).Create(gallery).Error
}

func (r *repository) UpdateGallery(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Gallery{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) DeleteGallery(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&Gallery{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *repository) IncrementViews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&Gallery{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
