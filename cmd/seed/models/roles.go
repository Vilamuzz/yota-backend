package models

import (
	"fmt"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) error {
	fmt.Println("Seeding roles...")
	roles := []account.Role{
		{ID: 1, Name: enum.RoleOrangTuaAsuh},
		{ID: 2, Name: enum.RoleChairman},
		{ID: 3, Name: enum.RoleSocialManager},
		{ID: 4, Name: enum.RoleFinance},
		{ID: 5, Name: enum.RoleAmbulanceManager},
		{ID: 6, Name: enum.RolePublicationManager},
		{ID: 7, Name: enum.RoleAmbulanceDriver},
		{ID: 8, Name: enum.RoleSuperadmin},
	}

	for _, role := range roles {
		if err := db.FirstOrCreate(&role, account.Role{Name: role.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed role '%s': %w", role.Name, err)
		}
	}
	return nil
}
