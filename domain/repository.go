package domain

import (
	"context"

	postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"
	"github.com/google/uuid"
)

type PostgreDBRepository interface {
	CreateOneUser(ctx context.Context, user *postgre_model.User) error
	FetchOneUser(ctx context.Context, options map[string]interface{}) (*postgre_model.User, error)
	CreatePasswordResetToken(ctx context.Context, token *postgre_model.PasswordResetToken) error
	FetchPasswordResetToken(ctx context.Context, token string) (*postgre_model.PasswordResetToken, error)
	UpdatePasswordResetToken(ctx context.Context, token *postgre_model.PasswordResetToken) error
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}
