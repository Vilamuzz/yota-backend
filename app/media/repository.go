package media

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	CreateMedia(ctx context.Context, media *Media) error
	CreateOneMedia(ctx context.Context, media *Media) error
	DeleteMedia(ctx context.Context, id string) error
	FetchMediaByEntity(ctx context.Context, entityID string, entityType string) ([]Media, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) CreateMedia(ctx context.Context, media *Media) error {
	return r.Conn.WithContext(ctx).Create(media).Error
}

func (r *repository) CreateOneMedia(ctx context.Context, media *Media) error {
	return r.Conn.WithContext(ctx).Create(media).Error
}

func (r *repository) DeleteMedia(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&Media{}).Error
}

func (r *repository) FetchMediaByEntity(ctx context.Context, entityID string, entityType string) ([]Media, error) {
	var mediaList []Media
	err := r.Conn.WithContext(ctx).
		Where("entity_id = ? AND entity_type = ?", entityID, entityType).
		Order("`order` ASC").
		Find(&mediaList).Error
	return mediaList, err
}
