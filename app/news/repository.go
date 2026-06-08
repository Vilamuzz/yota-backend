package news

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllNews(ctx context.Context, options map[string]interface{}) ([]News, error)
	CountNews(ctx context.Context, options map[string]interface{}) (int64, error)
	FindOneNews(ctx context.Context, options map[string]interface{}) (*News, error)
	CreateNews(ctx context.Context, news *News) error
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

var allowedNewsSortColumns = map[string]string{
	"title":        "title",
	"views":        "views",
	"published_at": "published_at",
	"created_at":   "created_at",
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
	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}

	// Determine sorting column and direction
	sortCol := "created_at"
	sortDir := "DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedNewsSortColumns[parts[0]]; valid {
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
	if err := query.Find(&newsList).Error; err != nil {
		return nil, err
	}
	return newsList, nil
}

func (r *repository) CountNews(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&News{}).Where("deleted_at IS NULL")

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

func (r *repository) FindOneNews(ctx context.Context, options map[string]interface{}) (*News, error) {
	var news News
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

func (r *repository) DeleteNews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *repository) IncrementViews(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&News{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
