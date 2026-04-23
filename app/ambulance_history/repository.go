package ambulance_history

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, ambulance AmbulanceHistory) error
	FindByID(ctx context.Context, id string) (AmbulanceHistory, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceHistory, error)
	Update(ctx context.Context, ambulance AmbulanceHistory) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, ambulance AmbulanceHistory) error {
	return r.Conn.Create(&ambulance).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (AmbulanceHistory, error) {
	var ambulance AmbulanceHistory
	if err := r.Conn.First(&ambulance, id).Error; err != nil {
		return AmbulanceHistory{}, err
	}
	return ambulance, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceHistory, error) {
	var ambulanceHistories []AmbulanceHistory
	query := r.Conn.WithContext(ctx)

	if ambulanceID, ok := options["ambulance_id"]; ok {
		query = query.Where("ambulance_id = ?", ambulanceID)
	}

	if ServiceCategory, ok := options["service_category"]; ok {
		query = query.Where("service_category = ?", ServiceCategory)
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
	if err := query.Find(&ambulanceHistories).Error; err != nil {
		return nil, err
	}
	return ambulanceHistories, nil
}

func (r *repository) Update(ctx context.Context, ambulance AmbulanceHistory) error {
	return r.Conn.Save(&ambulance).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.Conn.Delete(&AmbulanceHistory{}, id).Error
}
