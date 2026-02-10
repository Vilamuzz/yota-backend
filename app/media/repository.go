package media

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	DeleteEntityMedia(ctx context.Context, entityID string) error
	FetchEntityMedia(ctx context.Context, entityID string) ([]Media, error)
	CreateEntityMedia(ctx context.Context, entityID, entityType string, media []Media) error
	DeleteMediaByID(ctx context.Context, mediaID string) (*Media, error)
	FetchMediaByID(ctx context.Context, mediaID string) (*Media, error)
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

func (r *repository) CreateEntityMedia(ctx context.Context, entityID, entityType string, media []Media) error {
	// Set entity_id and entity_type for all media items
	for i := range media {
		media[i].EntityID = entityID
		media[i].EntityType = entityType
	}
	return r.Conn.WithContext(ctx).Create(&media).Error
}

func (r *repository) DeleteMediaByID(ctx context.Context, mediaID string) (*Media, error) {
	var media Media
	// Fetch the media first to return its info (needed for MinIO cleanup)
	if err := r.Conn.WithContext(ctx).Where("id = ?", mediaID).First(&media).Error; err != nil {
		return nil, err
	}
	// Delete the media
	if err := r.Conn.WithContext(ctx).Delete(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *repository) FetchMediaByID(ctx context.Context, mediaID string) (*Media, error) {
	var media Media
	if err := r.Conn.WithContext(ctx).Where("id = ?", mediaID).First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}
