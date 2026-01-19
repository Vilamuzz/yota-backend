package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateOneUser(ctx context.Context, user *User) error
	FetchOneUser(ctx context.Context, options map[string]interface{}) (*User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewPostgreRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) CreateOneUser(ctx context.Context, user *User) error {
	return r.Conn.WithContext(ctx).Create(user).Error
}

func (r *repository) FetchOneUser(ctx context.Context, options map[string]interface{}) (*User, error) {
	var user User
	if err := r.Conn.WithContext(ctx).Where(options).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}
