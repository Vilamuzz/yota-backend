package foundation_profile

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	FindFoundationProfile(ctx context.Context, options map[string]interface{}) (*FoundationProfile, error)
	CreateFoundationProfile(ctx context.Context, profile *FoundationProfile) error
	UpdateFoundationProfile(ctx context.Context, id string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FindFoundationProfile(ctx context.Context, options map[string]interface{}) (*FoundationProfile, error) {
	var profile FoundationProfile
	query := r.Conn.WithContext(ctx)

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *repository) CreateFoundationProfile(ctx context.Context, profile *FoundationProfile) error {
	return r.Conn.WithContext(ctx).Create(profile).Error
}

func (r *repository) UpdateFoundationProfile(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&FoundationProfile{}).Where("id = ?", id).Updates(updateData).Error
}