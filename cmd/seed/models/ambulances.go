package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedAmbulances(db *gorm.DB) error {
	fmt.Println("Seeding ambulances...")

	// Find the 5 driver accounts we seeded
	var drivers []account.Account
	emails := []string{
		"driver1@yota.com",
		"driver2@yota.com",
		"driver3@yota.com",
		"driver4@yota.com",
		"driver5@yota.com",
	}

	if err := db.Where("email IN ?", emails).Order("email ASC").Find(&drivers).Error; err != nil {
		return fmt.Errorf("failed to find driver accounts: %w", err)
	}

	if len(drivers) < 5 {
		return fmt.Errorf("not enough driver accounts found, expected 5, got %d", len(drivers))
	}

	now := time.Now()

	ambulances := []ambulance.Ambulance{
		{
			ID:          uuid.New(),
			DriverID:    drivers[0].ID,
			Image:       "https://images.unsplash.com/photo-1612574935301-af13ccce9258?q=80&w=1470&auto=format&fit=crop&w=800&q=80",
			PlateNumber: "B 1234 SAA",
			Status:      ambulance.AmbulanceStatusAvailable,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			DriverID:    drivers[1].ID,
			Image:       "https://images.unsplash.com/photo-1705264895993-c544cf74a0c7?q=80&w=1470&auto=format&fit=crop&w=800&q=80",
			PlateNumber: "B 5678 SBB",
			Status:      ambulance.AmbulanceStatusInUse,
			CreatedAt:   now.Add(-time.Hour * 12),
			UpdatedAt:   now.Add(-time.Hour * 12),
		},
		{
			ID:          uuid.New(),
			DriverID:    drivers[2].ID,
			Image:       "https://images.unsplash.com/photo-1587745416684-47953f16f02f?q=80&w=1408&auto=format&fit=crop&w=800&q=80",
			PlateNumber: "B 9012 SCC",
			Status:      ambulance.AmbulanceStatusAvailable,
			CreatedAt:   now.Add(-time.Hour * 24),
			UpdatedAt:   now.Add(-time.Hour * 24),
		},
		{
			ID:          uuid.New(),
			DriverID:    drivers[3].ID,
			Image:       "https://images.unsplash.com/photo-1697952431905-9c8d169d9d2b?w=600&auto=format&fit=crop&w=800&q=80",
			PlateNumber: "B 3456 SDD",
			Status:      ambulance.AmbulanceStatusMaintenance,
			CreatedAt:   now.Add(-time.Hour * 36),
			UpdatedAt:   now.Add(-time.Hour * 36),
		},
		{
			ID:          uuid.New(),
			DriverID:    drivers[4].ID,
			Image:       "https://images.unsplash.com/photo-1633094832342-b6018253d4e6?q=80&w=1470&auto=format&fit=crop&w=800&q=80",
			PlateNumber: "B 7890 SEE",
			Status:      ambulance.AmbulanceStatusAvailable,
			CreatedAt:   now.Add(-time.Hour * 48),
			UpdatedAt:   now.Add(-time.Hour * 48),
		},
	}

	for _, amb := range ambulances {
		var existing ambulance.Ambulance
		err := db.Where("plate_number = ?", amb.PlateNumber).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&amb).Error; err != nil {
					return fmt.Errorf("failed to create ambulance with plate number %s: %w", amb.PlateNumber, err)
				}
				fmt.Printf("✓ Created ambulance: %s (Driver: %s)\n", amb.PlateNumber, amb.DriverID)
			} else {
				return fmt.Errorf("error checking existing ambulance: %w", err)
			}
		} else {
			fmt.Printf("⚠ Ambulance '%s' already exists, skipping...\n", amb.PlateNumber)
		}
	}

	return nil
}
