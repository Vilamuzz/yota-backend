package auth

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	FetchPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	UpdatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.Conn.WithContext(ctx).Create(token).Error
}

func (r *repository) FetchPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	var resetToken PasswordResetToken
	if err := r.Conn.WithContext(ctx).Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).First(&resetToken).Error; err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *repository) UpdatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.Conn.WithContext(ctx).Save(token).Error
}
