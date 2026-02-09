package gallery

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FetchAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error)
	FetchGalleryByID(ctx context.Context, id string) (*Gallery, error)
	CreateOneGallery(ctx context.Context, gallery *Gallery) error
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

func (r *repository) FetchAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error) {
	var galleries []Gallery
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

	if err := query.Find(&galleries).Error; err != nil {
		return nil, err
	}

	return galleries, nil
}

func (r *repository) FetchGalleryByID(ctx context.Context, id string) (*Gallery, error) {
	var gallery Gallery
	if err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&gallery).Error; err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (r *repository) CreateOneGallery(ctx context.Context, gallery *Gallery) error {
	return r.Conn.WithContext(ctx).Create(gallery).Error
}

func (r *repository) UpdateGallery(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Gallery{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) DeleteGallery(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&Gallery{}).Error
}

func (r *repository) IncrementViews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&Gallery{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
