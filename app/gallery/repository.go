package gallery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error)
	CountGalleries(ctx context.Context, options map[string]interface{}) (int64, error)
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

var allowedGallerySortColumns = map[string]string{
	"title":      "title",
	"views":      "views",
	"created_at": "created_at",
}

func (r *repository) FindAllGalleries(ctx context.Context, options map[string]interface{}) ([]Gallery, error) {
	var galleries []Gallery
	query := r.Conn.WithContext(ctx).Preload("Media").Where("deleted_at IS NULL")

	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}

	// Determine sorting column and direction
	sortCol := "created_at"
	sortDir := "DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedGallerySortColumns[parts[0]]; valid {
				sortCol = col
				if len(parts) == 2 && (parts[1] == "asc" || parts[1] == "desc") {
					sortDir = strings.ToUpper(parts[1])
				}
			}
		}
	}

	query = query.Order(fmt.Sprintf("%s %s", sortCol, sortDir))

	limit := 10
	if l, ok := options["limit"]; ok && l.(int) > 0 {
		limit = l.(int)
	}

	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&galleries).Error; err != nil {
		return nil, err
	}
	return galleries, nil
}

func (r *repository) CountGalleries(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&Gallery{}).Where("deleted_at IS NULL")

	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}

	err := query.Count(&total).Error
	return total, err
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
