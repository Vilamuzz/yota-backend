package user

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllRoles(ctx context.Context) ([]Role, error)
	FindRoleByID(ctx context.Context, roleID int8) (*Role, error)
	Create(ctx context.Context, user *User) error
	FindOne(ctx context.Context, options map[string]interface{}) (*User, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	Update(ctx context.Context, userID string, updateData map[string]interface{}) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	if err := r.Conn.WithContext(ctx).Where("role != ?", "superadmin").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) FindRoleByID(ctx context.Context, roleID int8) (*Role, error) {
	var role Role
	if err := r.Conn.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.Conn.WithContext(ctx).Create(user).Error
}

func (r *repository) FindOne(ctx context.Context, options map[string]interface{}) (*User, error) {
	var user User
	if err := r.Conn.WithContext(ctx).Where(options).Preload("Role").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]User, error) {
	var users []User
	query := r.Conn.WithContext(ctx).Preload("Role").Where("role_id != ?", 8)

	// Apply filters
	if roleID, ok := options["role_id"]; ok && roleID != 0 {
		query = query.Where("role_id = ?", roleID)
	}

	if status, ok := options["status"]; ok && status != nil {
		query = query.Where("status = ?", status)
	}

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search.(string)+"%", "%"+search.(string)+"%")
	}

	// Apply cursor-based pagination
	if cursor, ok := options["next_cursor"]; ok && cursor != "" {
		cursorData, err := pkg.DecodeCursor(cursor.(string))
		if err == nil {
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID).
				Order("created_at ASC, id ASC")
		}
	}

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	query = query.Limit(limit + 1)

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *repository) Update(ctx context.Context, userID string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(updateData).Error
}

func (r *repository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	return r.Conn.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("email_verified", true).Error
}
