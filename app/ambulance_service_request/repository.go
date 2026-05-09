package ambulance_service_request

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, ambulanceServiceRequest AmbulanceServiceRequest) error
	FindByID(ctx context.Context, id string) (AmbulanceServiceRequest, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceServiceRequest, error)
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, ambulanceServiceRequest AmbulanceServiceRequest) error {
	return r.Conn.Create(&ambulanceServiceRequest).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (AmbulanceServiceRequest, error) {
	var ambulanceServiceRequest AmbulanceServiceRequest
	if err := r.Conn.First(&ambulanceServiceRequest, id).Error; err != nil {
		return AmbulanceServiceRequest{}, err
	}
	return ambulanceServiceRequest, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceServiceRequest, error) {
	var ambulanceServiceRequests []AmbulanceServiceRequest
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
	if err := query.Find(&ambulanceServiceRequests).Error; err != nil {
		return nil, err
	}
	return ambulanceServiceRequests, nil
}

func (r *repository) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.Model(&AmbulanceServiceRequest{}).Where("id = ?", id).Updates(updateData).Error
}
