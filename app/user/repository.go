package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FetchRoleByID(ctx context.Context, roleID int8) (*Role, error)
	CreateOneUser(ctx context.Context, user *User) error
	FetchOneUser(ctx context.Context, options map[string]interface{}) (*User, error)
	FetchListUsers(ctx context.Context, options map[string]interface{}) ([]User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateUser(ctx context.Context, userID string, updateData map[string]interface{}) error
	VerifyUserEmail(ctx context.Context, userID uuid.UUID) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FetchRoleByID(ctx context.Context, roleID int8) (*Role, error) {
	var role Role
	if err := r.Conn.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) CreateOneUser(ctx context.Context, user *User) error {
	return r.Conn.WithContext(ctx).Create(user).Error
}

func (r *repository) FetchOneUser(ctx context.Context, options map[string]interface{}) (*User, error) {
	var user User
	if err := r.Conn.WithContext(ctx).Where(options).Preload("Role").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FetchListUsers(ctx context.Context, options map[string]interface{}) ([]User, error) {
	var users []User
	if err := r.Conn.WithContext(ctx).Where(options).Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *repository) UpdateUser(ctx context.Context, userID string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(updateData).Error
}

func (r *repository) VerifyUserEmail(ctx context.Context, userID uuid.UUID) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("email_verified", true).Error
}
