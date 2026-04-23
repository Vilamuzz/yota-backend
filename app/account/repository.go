package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllAccounts(ctx context.Context, options map[string]interface{}) ([]Account, error)
	FindOneAccount(ctx context.Context, options map[string]interface{}) (*Account, error)
	CreateAccount(ctx context.Context, account *Account) error
	UpdateAccount(ctx context.Context, accountID string, updateData map[string]interface{}) error

	CountActiveAccountRoles(ctx context.Context, accountID string) (int64, error)
	FindOneAccountRole(ctx context.Context, accountID string, roleID int) (*AccountRole, error)
	CreateAccountRole(ctx context.Context, accountRole *AccountRole) error
	UpdateAccountRole(ctx context.Context, accountID string, roleID int, updateData map[string]interface{}) error

	FindAllRoles(ctx context.Context) ([]Role, error)
	FindOneRole(ctx context.Context, roleID int) (*Role, error)
	UpdateFullProfile(ctx context.Context, accountID string, updateAccount, updateProfile map[string]interface{}, defaultRoleID int) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllAccounts(ctx context.Context, options map[string]interface{}) ([]Account, error) {
	var accounts []Account
	query := r.Conn.WithContext(ctx).
		Preload("UserProfile").
		Preload("AccountRoles.Role").
		Joins("LEFT JOIN account_roles ON account_roles.account_id = accounts.id").
		Group("accounts.id")

	if excludeSuperadmin, ok := options["exclude_superadmin"]; ok && excludeSuperadmin.(bool) {
		query = query.Where("accounts.id NOT IN (SELECT account_id FROM account_roles WHERE role_id = ?)", 8)
	}
	if roleID, ok := options["role_id"]; ok && roleID != 0 {
		query = query.Where("account_roles.role_id = ?", roleID)
	}
	if isBanned, ok := options["is_banned"]; ok {
		query = query.Where("accounts.is_banned = ?", isBanned.(bool))
	}
	if search, ok := options["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Joins("UserProfile").Where("user_profiles.username LIKE ? OR accounts.email LIKE ?", searchStr, searchStr)
	}
	sortOrder := enum.SortOrderDesc
	if val, ok := options["sort_order"].(enum.SortOrderEnum); ok && val != "" {
		sortOrder = val
	}
	if nextCursor, ok := options["next_cursor"]; ok && nextCursor != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			if sortOrder == enum.SortOrderDesc {
				query = query.Where("accounts.created_at < ? OR (accounts.created_at = ? AND accounts.id < ?)",
					cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
			} else {
				query = query.Where("accounts.created_at > ? OR (accounts.created_at = ? AND accounts.id > ?)",
					cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
			}
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			if sortOrder == enum.SortOrderDesc {
				query = query.Where("accounts.created_at > ? OR (accounts.created_at = ? AND accounts.id > ?)",
					cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
			} else {
				query = query.Where("accounts.created_at < ? OR (accounts.created_at = ? AND accounts.id < ?)",
					cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
			}
		}

	}

	if _, ok := options["prev_cursor"]; ok {
		if sortOrder == enum.SortOrderDesc {
			query = query.Order("accounts.created_at ASC, accounts.id ASC")
		} else {
			query = query.Order("accounts.created_at DESC, accounts.id DESC")
		}
	} else {
		query = query.Order(fmt.Sprintf("accounts.created_at %s, accounts.id %s", sortOrder, sortOrder))
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *repository) FindOneAccount(ctx context.Context, options map[string]interface{}) (*Account, error) {
	var account Account

	query := r.Conn.WithContext(ctx).Preload("UserProfile").Preload("AccountRoles.Role")
	if email, ok := options["email"]; ok {
		query = query.Where("accounts.email = ?", email)
	}
	if id, ok := options["id"]; ok {
		query = query.Where("accounts.id = ?", id)
	}
	if username, ok := options["username"]; ok {
		query = query.Joins("UserProfile").Where("user_profiles.username = ?", username)
	}
	if phone, ok := options["phone"]; ok {
		query = query.Joins("UserProfile").Where("user_profiles.phone = ?", phone)
	}
	if err := query.First(&account).Error; err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *repository) CreateAccount(ctx context.Context, account *Account) error {
	return r.Conn.WithContext(ctx).Create(account).Error
}

func (r *repository) UpdateAccount(ctx context.Context, accountID string, updateData map[string]interface{}) error {
	result := r.Conn.WithContext(ctx).
		Model(&Account{}).
		Where("id = ?", accountID).
		Updates(updateData)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no account found to update, or data was unchanged")
	}

	return nil
}

func (r *repository) CountActiveAccountRoles(ctx context.Context, accountID string) (int64, error) {
	var count int64
	query := r.Conn.WithContext(ctx).Where("is_active = ? AND account_id = ?", true, accountID)
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repository) FindOneAccountRole(ctx context.Context, accountID string, roleID int) (*AccountRole, error) {
	var accountRole AccountRole
	if err := r.Conn.WithContext(ctx).Where("account_id = ? AND role_id = ?", accountID, roleID).First(&accountRole).Error; err != nil {
		return nil, err
	}

	return &accountRole, nil
}

func (r *repository) CreateAccountRole(ctx context.Context, accountRole *AccountRole) error {
	return r.Conn.WithContext(ctx).Create(accountRole).Error
}

func (r *repository) UpdateAccountRole(ctx context.Context, accountID string, roleID int, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&AccountRole{}).Where("account_id = ? AND role_id = ?", accountID, roleID).Updates(updateData).Error
}

func (r *repository) FindAllRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	if err := r.Conn.WithContext(ctx).Where("role != ?", "superadmin").Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *repository) FindOneRole(ctx context.Context, roleID int) (*Role, error) {
	var role Role
	if err := r.Conn.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *repository) UpdateFullProfile(ctx context.Context, accountID string, updateAccount, updateProfile map[string]interface{}, defaultRoleID int) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(updateAccount) > 0 {
			if err := tx.Model(&Account{}).Where("id = ?", accountID).Updates(updateAccount).Error; err != nil {
				return err
			}
		}

		if len(updateProfile) > 0 {
			if err := tx.Model(&UserProfile{}).Where("account_id = ?", accountID).Updates(updateProfile).Error; err != nil {
				return err
			}
		}

		if defaultRoleID != 0 {
			if err := tx.Model(&AccountRole{}).Where("account_id = ?", accountID).Update("is_default", false).Error; err != nil {
				return err
			}
			if err := tx.Model(&AccountRole{}).Where("account_id = ? AND role_id = ?", accountID, defaultRoleID).Update("is_default", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
