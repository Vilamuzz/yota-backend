package main

import (
	"fmt"

	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"gorm.io/gorm"
)

func seedRoles(db *gorm.DB) error {
	fmt.Println("Seeding roles...")
	roles := []user.Role{
		{ID: 1, Role: string(enum.RoleUser)},
		{ID: 2, Role: string(enum.RoleChairman)},
		{ID: 3, Role: string(enum.RoleSocialManager)},
		{ID: 4, Role: string(enum.RoleFinance)},
		{ID: 5, Role: string(enum.RoleAmbulanceManager)},
		{ID: 6, Role: string(enum.RolePublicationManager)},
		{ID: 7, Role: string(enum.RoleAmbulanceDriver)},
		{ID: 8, Role: string(enum.RoleSuperadmin)},
	}

	for _, role := range roles {
		if err := db.FirstOrCreate(&role, user.Role{Role: role.Role}).Error; err != nil {
			return fmt.Errorf("failed to seed role '%s': %w", role.Role, err)
		}
	}
	return nil
}
