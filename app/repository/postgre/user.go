package postgre_repository

import (
	"context"
	"time"

	postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"
	"github.com/google/uuid"
)

func (r *postgreDbRepo) CreateOneUser(ctx context.Context, user *postgre_model.User) error {
	return r.Conn.WithContext(ctx).Create(user).Error
}

func (r *postgreDbRepo) FetchOneUser(ctx context.Context, options map[string]interface{}) (*postgre_model.User, error) {
	var user postgre_model.User
	if err := r.Conn.WithContext(ctx).Where(options).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgreDbRepo) CreatePasswordResetToken(ctx context.Context, token *postgre_model.PasswordResetToken) error {
	return r.Conn.WithContext(ctx).Create(token).Error
}

func (r *postgreDbRepo) FetchPasswordResetToken(ctx context.Context, token string) (*postgre_model.PasswordResetToken, error) {
	var resetToken postgre_model.PasswordResetToken
	if err := r.Conn.WithContext(ctx).Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).First(&resetToken).Error; err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *postgreDbRepo) UpdatePasswordResetToken(ctx context.Context, token *postgre_model.PasswordResetToken) error {
	return r.Conn.WithContext(ctx).Save(token).Error
}

func (r *postgreDbRepo) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.Conn.WithContext(ctx).Model(&postgre_model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}
