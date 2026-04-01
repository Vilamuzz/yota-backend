package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `json:"id" gorm:"primary_key"`
	Username      string    `json:"username" gorm:"unique;not null"`
	Email         string    `json:"email" gorm:"unique;not null"`
	Password      string    `json:"password" gorm:"not null"`
	Status        bool      `json:"status" gorm:"type:boolean;not null;default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	DefaultRoleID int8   `json:"default_role_id" gorm:"not null"`
	DefaultRole   Role   `json:"default_role" gorm:"foreignKey:DefaultRoleID"`
	Roles         []Role `json:"roles" gorm:"many2many:user_roles;"`
}

type Role struct {
	ID   int8   `json:"id" gorm:"primary_key"`
	Role string `json:"role" gorm:"type:varchar(20);not null;unique"`
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id" gorm:"primary_key"`
	RoleID int8      `json:"role_id" gorm:"primary_key"`
}
