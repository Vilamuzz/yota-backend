package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"primary_key"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"password" gorm:"not null"`
	Role      Role      `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	Status    Status    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role string

const (
	RoleUser               Role = "user"
	RoleChairman           Role = "chairman"
	RoleSocialManager      Role = "social_manager"
	RoleFinance            Role = "finance"
	RoleAmbulanceManager   Role = "ambulance_manager"
	RolePublicationManager Role = "publication_manager"
	RoleSuperadmin         Role = "superadmin"
)

type Status string

const (
	StatusActive Status = "active"
	StatusBanned Status = "nonactive"
)
