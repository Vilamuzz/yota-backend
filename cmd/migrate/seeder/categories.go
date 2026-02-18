package main

import (
	"fmt"
	"log"

	"github.com/Vilamuzz/yota-backend/app/media"
	"gorm.io/gorm"
)

func seedCategories(db *gorm.DB) error {
	fmt.Println("Seeding categories...")

	categories := []media.CategoryMedia{
		{
			ID:       1,
			Category: "Photography",
		},
		{
			ID:       2,
			Category: "Painting",
		},
		{
			ID:       3,
			Category: "Sculpture",
		},
		{
			ID:       4,
			Category: "Digital",
		},
		{
			ID:       5,
			Category: "Mixed",
		},
	}

	for _, c := range categories {
		var existingCategory media.CategoryMedia
		if err := db.Where("Category = ?", c.Category).First(&existingCategory).Error; err == nil {
			fmt.Printf("⚠ Category %s already exists, skipping...\n", c.Category)
			continue
		}

		if err := db.Create(&c).Error; err != nil {
			log.Printf("Warning: Failed to create category %s: %v", c.Category, err)
			continue
		}
		fmt.Printf("✓ Created category: %s\n", c.Category)
	}

	return nil
}
