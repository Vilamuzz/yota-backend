package postgre_repository

import (
	"context"

	postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"
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
