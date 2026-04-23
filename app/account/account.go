package account

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/google/uuid"
)

type Account struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey"`
	Email         string    `json:"email" gorm:"unique;not null"`
	Password      string    `json:"password" gorm:"not null"`
	IsBanned      bool      `json:"is_banned" gorm:"type:boolean;not null;default:false"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	UserProfile  UserProfile   `json:"user_profile" gorm:"foreignKey:AccountID;references:ID"`
	AccountRoles []AccountRole `json:"account_roles" gorm:"foreignKey:AccountID;references:ID"`
}

type UserProfile struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID      uuid.UUID `json:"account_id" gorm:"unique;not null"`
	Username       string    `json:"username" gorm:"unique"`
	Phone          string    `json:"phone" gorm:"unique"`
	Address        string    `json:"address"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Role struct {
	ID   int           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name enum.RoleName `json:"name" gorm:"type:varchar(20);not null;unique"`
}

type AccountRole struct {
	AccountID uuid.UUID `json:"account_id" gorm:"primaryKey"`
	RoleID    int       `json:"role_id" gorm:"primaryKey"`
	IsDefault bool      `json:"is_default" gorm:"default:false"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`

	Account Account `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	Role    Role    `json:"role" gorm:"foreignKey:RoleID;references:ID"`
}

const ProtectedSuperAdminRoleID = 8
const OrangTuaAsuhRoleID = 1
