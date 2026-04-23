package ambulance

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, ambulance Ambulance) error
	FindByID(ctx context.Context, id string) (Ambulance, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]Ambulance, error)
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, ambulance Ambulance) error {
	return r.Conn.WithContext(ctx).Create(&ambulance).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (Ambulance, error) {
	var ambulance Ambulance
	if err := r.Conn.WithContext(ctx).First(&ambulance, id).Error; err != nil {
		return Ambulance{}, err
	}
	return ambulance, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Ambulance, error) {
	var ambulances []Ambulance
	query := r.Conn.WithContext(ctx)

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("plate_number LIKE ?", "%"+search.(string)+"%")
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
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID).
				Order("created_at ASC, id ASC")
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
	if err := query.Find(&ambulances).Error; err != nil {
		return nil, err
	}
	return ambulances, nil
}

func (r *repository) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Ambulance{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Delete(&Ambulance{}, id).Error
}
