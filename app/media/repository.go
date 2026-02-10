package media

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	DeleteEntityMedia(ctx context.Context, entityID string) error
	FetchEntityMedia(ctx context.Context, entityID string) ([]Media, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) DeleteEntityMedia(ctx context.Context, entityID string) error {
	return r.Conn.WithContext(ctx).Where("entity_id = ?", entityID).Delete(&Media{}).Error
}

func (r *repository) FetchEntityMedia(ctx context.Context, entityID string) ([]Media, error) {
	var media []Media
	if err := r.Conn.WithContext(ctx).Where("entity_id = ?", entityID).Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}
