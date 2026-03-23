package prayer

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, prayer *Prayer) error
	Update(ctx context.Context, prayer *Prayer) error
	FindByID(ctx context.Context, id string) (*Prayer, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]Prayer, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, prayer *Prayer) error {
	return r.Conn.Create(prayer).Error
}

func (r *repository) Update(ctx context.Context, prayer *Prayer) error {
	return r.Conn.Save(prayer).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Prayer, error) {
	var prayer Prayer
	if err := r.Conn.First(&prayer, id).Error; err != nil {
		return nil, err
	}
	return &prayer, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Prayer, error) {
	var prayers []Prayer
	query := r.Conn.WithContext(ctx).Preload("User")

	if donationID, ok := options["donation_id"]; ok {
		query = query.Where("donation_id = ?", donationID)
	}
	if reported, ok := options["reported"]; ok {
		if reported.(bool) {
			query = query.Where("report_count > ?", 0)
		}
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

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&prayers).Error; err != nil {
		return nil, err
	}
	return prayers, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.Conn.Delete(&Prayer{}, id).Error
}
