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
	IsBanned      bool      `json:"isBanned" gorm:"type:boolean;not null;default:false"`
	EmailVerified bool      `json:"emailVerified" gorm:"default:false"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	UserProfile  UserProfile   `json:"userProfile" gorm:"foreignKey:AccountID;references:ID"`
	AccountRoles []AccountRole `json:"accountRoles" gorm:"foreignKey:AccountID;references:ID"`
}

type UserProfile struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey"`
	AccountID      uuid.UUID `json:"accountId" gorm:"unique;not null"`
	Username       string    `json:"username" gorm:"unique"`
	Phone          *string   `json:"phone" gorm:"unique"`
	Address        string    `json:"address"`
	ProfilePicture string    `json:"profilePicture"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Role struct {
	ID   int           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name enum.RoleName `json:"name" gorm:"type:varchar(30);not null;unique"`
}

type AccountRole struct {
	AccountID uuid.UUID `json:"accountId" gorm:"primaryKey"`
	RoleID    int       `json:"roleId" gorm:"primaryKey"`
	IsDefault bool      `json:"isDefault" gorm:"default:false"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`

	Account Account `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	Role    Role    `json:"role" gorm:"foreignKey:RoleID;references:ID"`
}

const ProtectedSuperAdminRoleID = 8
const OrangTuaAsuhRoleID = 1
