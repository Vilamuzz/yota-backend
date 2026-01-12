package domain

import (
	"context"

	postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"
)

type PostgreDBRepository interface {
	CreateOneUser(ctx context.Context, user *postgre_model.User) error
	FetchOneUser(ctx context.Context, options map[string]interface{}) (*postgre_model.User, error)
}
