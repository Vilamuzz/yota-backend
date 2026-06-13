package ambulance

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllAmbulances(ctx context.Context, options map[string]interface{}) ([]Ambulance, error)
	FindOneAmbulance(ctx context.Context, options map[string]interface{}) (*Ambulance, error)
	CreateAmbulance(ctx context.Context, ambulance *Ambulance) error
	UpdateAmbulance(ctx context.Context, id string, updateData map[string]interface{}) error
	DeleteAmbulance(ctx context.Context, id string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FindAllAmbulances(ctx context.Context, options map[string]interface{}) ([]Ambulance, error) {
	var ambulances []Ambulance
	query := r.Conn.WithContext(ctx).
		Joins("LEFT JOIN accounts ON accounts.id = ambulances.driver_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.account_id = accounts.id").
		Preload("Driver.UserProfile").
		Where("ambulances.deleted_at IS NULL").
		Select("ambulances.*")

	if search, ok := options["search"]; ok && search != "" {
		searchQuery := "%" + search.(string) + "%"
		query = query.Where("ambulances.plate_number ILIKE ? OR user_profiles.username ILIKE ?", searchQuery, searchQuery)
	}

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("ambulances.status = ?", status)
	}

	if driverID, ok := options["driver_id"]; ok && driverID != "" {
		query = query.Where("ambulances.driver_id = ?", driverID)
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("ambulances.created_at < ? OR (ambulances.created_at = ? AND ambulances.id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("ambulances.created_at > ? OR (ambulances.created_at = ? AND ambulances.id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	if _, isPrev := options["prev_cursor"]; isPrev {
		query = query.Order("ambulances.created_at ASC, ambulances.id ASC")
	} else {
		query = query.Order("ambulances.created_at DESC, ambulances.id DESC")
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

func (r *repository) FindOneAmbulance(ctx context.Context, options map[string]interface{}) (*Ambulance, error) {
	var ambulance Ambulance
	query := r.Conn.WithContext(ctx).Preload("Driver.UserProfile").Where("deleted_at IS NULL")

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if plateNumber, ok := options["plate_number"]; ok && plateNumber != "" {
		query = query.Where("plate_number = ?", plateNumber)
	}
	if driverID, ok := options["driver_id"]; ok && driverID != "" {
		query = query.Where("driver_id = ?", driverID)
	}

	if err := query.First(&ambulance).Error; err != nil {
		return nil, err
	}
	return &ambulance, nil
}

func (r *repository) CreateAmbulance(ctx context.Context, ambulance *Ambulance) error {
	return r.Conn.WithContext(ctx).Create(ambulance).Error
}

func (r *repository) UpdateAmbulance(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Ambulance{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) DeleteAmbulance(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&Ambulance{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
