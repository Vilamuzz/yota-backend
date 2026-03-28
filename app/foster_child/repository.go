package foster_child

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindOne(ctx context.Context, options map[string]interface{}) (*FosterChild, error) {
	var fosterChild FosterChild
	collectedFundSubquery := r.Conn.Table("foster_child_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("foster_child_id = foster_children.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("foster_children.*, (?) as collected_fund", collectedFundSubquery)
	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("slug = ?", slug)
	}
	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.First(&fosterChild).Error; err != nil {
		return nil, err
	}
	return &fosterChild, nil
}
