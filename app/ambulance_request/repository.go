package ambulance_request

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, ambulanceRequest AmbulanceRequest) error
	FindByID(ctx context.Context, id string) (AmbulanceRequest, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceRequest, error)
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, ambulanceRequest AmbulanceRequest) error {
	return r.Conn.Create(&ambulanceRequest).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (AmbulanceRequest, error) {
	var ambulanceRequest AmbulanceRequest
	if err := r.Conn.First(&ambulanceRequest, id).Error; err != nil {
		return AmbulanceRequest{}, err
	}
	return ambulanceRequest, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceRequest, error) {
	var ambulanceRequests []AmbulanceRequest
	query := r.Conn.WithContext(ctx)

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
	if err := query.Find(&ambulanceRequests).Error; err != nil {
		return nil, err
	}
	return ambulanceRequests, nil
}

func (r *repository) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.Model(&AmbulanceRequest{}).Where("id = ?", id).Updates(updateData).Error
}
