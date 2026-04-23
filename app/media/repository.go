package media

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	DeleteEntityMedia(ctx context.Context, entityID, entityType string) error
	FetchEntityMedia(ctx context.Context, entityID, entityType string) ([]Media, error)
	CreateEntityMedia(ctx context.Context, entityID, entityType string, media []Media) error
	DeleteMediaByID(ctx context.Context, mediaID string) (*Media, error)
	FetchMediaByID(ctx context.Context, mediaID string) (*Media, error)
	UpdateMediaByID(ctx context.Context, mediaID string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FetchEntityMedia(ctx context.Context, entityID, entityType string) ([]Media, error) {
	var media []Media
	query := r.Conn.WithContext(ctx)
	switch entityType {
	case "news":
		query = query.Where("news_id = ?", entityID)
	case "gallery":
		query = query.Where("gallery_id = ?", entityID)
	default:
		return nil, nil
	}

	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (r *repository) FetchMediaByID(ctx context.Context, mediaID string) (*Media, error) {
	var media Media
	if err := r.Conn.WithContext(ctx).Where("id = ?", mediaID).First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *repository) CreateEntityMedia(ctx context.Context, entityID, entityType string, media []Media) error {
	for i := range media {
		switch entityType {
		case "news":
			media[i].NewsID = uuid.MustParse(entityID)
		case "gallery":
			media[i].GalleryID = uuid.MustParse(entityID)
		}
	}
	return r.Conn.WithContext(ctx).Create(&media).Error
}

func (r *repository) UpdateMediaByID(ctx context.Context, mediaID string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Media{}).Where("id = ?", mediaID).Updates(updateData).Error
}

func (r *repository) DeleteEntityMedia(ctx context.Context, entityID, entityType string) error {
	query := r.Conn.WithContext(ctx)
	switch entityType {
	case "news":
		query = query.Where("news_id = ?", entityID)
	case "gallery":
		query = query.Where("gallery_id = ?", entityID)
	default:
		return nil
	}
	return query.Delete(&Media{}).Error
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
