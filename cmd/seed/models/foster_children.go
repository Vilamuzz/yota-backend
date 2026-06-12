package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedFosterChildren(db *gorm.DB) error {
	fmt.Println("Seeding foster children...")

	now := time.Now()

	children := []foster_children.FosterChildren{
		{
			ID:             uuid.New(),
			Name:           "Ahmad Fadillah",
			ProfilePicture: "https://images.unsplash.com/photo-1595454233719-743cc1eb3602?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Male,
			IsGraduated:    false,
			Category:       foster_children.CategoryOrphan,
			BirthDate:      time.Date(2010, 5, 14, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Jakarta",
			SchoolName:     "SMPN 1 Jakarta",
			EducationLevel: 8,
			Address:        "Jl. Merdeka No. 10, Jakarta Selatan",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			Name:           "Siti Aminah",
			ProfilePicture: "https://images.unsplash.com/photo-1544640808-32cb4fba5901?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Female,
			IsGraduated:    false,
			Category:       foster_children.CategoryFatherless,
			BirthDate:      time.Date(2012, 8, 22, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Bandung",
			SchoolName:     "SDN Cibeureum",
			EducationLevel: 6,
			Address:        "Jl. Diponegoro No. 45, Bandung",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 24),
			UpdatedAt:      now.Add(-time.Hour * 24),
		},
		{
			ID:             uuid.New(),
			Name:           "Budi Santoso",
			ProfilePicture: "https://images.unsplash.com/photo-1508344928928-7157b83d1667?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Male,
			IsGraduated:    false,
			Category:       foster_children.CategoryMotherless,
			BirthDate:      time.Date(2008, 11, 2, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Surabaya",
			SchoolName:     "SMAN 5 Surabaya",
			EducationLevel: 10,
			Address:        "Jl. Pahlawan No. 12, Surabaya",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 48),
			UpdatedAt:      now.Add(-time.Hour * 48),
		},
		{
			ID:             uuid.New(),
			Name:           "Nurul Hidayati",
			ProfilePicture: "https://images.unsplash.com/photo-1621360064228-4ce6da55787f?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Female,
			IsGraduated:    false,
			Category:       foster_children.CategoryOrphan,
			BirthDate:      time.Date(2014, 2, 18, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Semarang",
			SchoolName:     "SDN Pandean Lamper",
			EducationLevel: 4,
			Address:        "Jl. Gajah Mada No. 8, Semarang",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 72),
			UpdatedAt:      now.Add(-time.Hour * 72),
		},
		{
			ID:             uuid.New(),
			Name:           "Rizky Maulana",
			ProfilePicture: "https://images.unsplash.com/photo-1581452902341-b845fec0d53c?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Male,
			IsGraduated:    false,
			Category:       foster_children.CategoryFatherless,
			BirthDate:      time.Date(2009, 9, 30, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Yogyakarta",
			SchoolName:     "SMPN 8 Yogyakarta",
			EducationLevel: 9,
			Address:        "Jl. Kaliurang KM 5, Sleman",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 96),
			UpdatedAt:      now.Add(-time.Hour * 96),
		},
		{
			ID:             uuid.New(),
			Name:           "Aulia Rahman",
			ProfilePicture: "https://images.unsplash.com/photo-1580129994689-d4ff51a56112?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Female,
			IsGraduated:    true,
			Category:       foster_children.CategoryOrphan,
			BirthDate:      time.Date(2005, 4, 12, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Malang",
			SchoolName:     "SMAN 3 Malang",
			EducationLevel: 12,
			Address:        "Jl. Ijen No. 25, Malang",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 120),
			UpdatedAt:      now.Add(-time.Hour * 120),
		},
		{
			ID:             uuid.New(),
			Name:           "Fikri Hakim",
			ProfilePicture: "https://images.unsplash.com/photo-1542385151-efd9000785a0?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Male,
			IsGraduated:    false,
			Category:       foster_children.CategoryMotherless,
			BirthDate:      time.Date(2013, 7, 5, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Medan",
			SchoolName:     "SDN 060911 Medan",
			EducationLevel: 5,
			Address:        "Jl. Jamin Ginting No. 100, Medan",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 144),
			UpdatedAt:      now.Add(-time.Hour * 144),
		},
		{
			ID:             uuid.New(),
			Name:           "Nadia Salsabila",
			ProfilePicture: "https://images.unsplash.com/photo-1517855325881-22e6b7d5a57d?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Female,
			IsGraduated:    false,
			Category:       foster_children.CategoryFatherless,
			BirthDate:      time.Date(2011, 1, 20, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Padang",
			SchoolName:     "SMPN 2 Padang",
			EducationLevel: 7,
			Address:        "Jl. Sudirman No. 50, Padang",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 168),
			UpdatedAt:      now.Add(-time.Hour * 168),
		},
		{
			ID:             uuid.New(),
			Name:           "Ilham Pratama",
			ProfilePicture: "https://images.unsplash.com/photo-1629813296837-7756e01a89ea?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Male,
			IsGraduated:    false,
			Category:       foster_children.CategoryOrphan,
			BirthDate:      time.Date(2007, 12, 10, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Palembang",
			SchoolName:     "SMAN 1 Palembang",
			EducationLevel: 11,
			Address:        "Jl. Veteran No. 33, Palembang",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 192),
			UpdatedAt:      now.Add(-time.Hour * 192),
		},
		{
			ID:             uuid.New(),
			Name:           "Dewi Lestari",
			ProfilePicture: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?auto=format&fit=crop&w=800&q=80",
			Gender:         foster_children.Female,
			IsGraduated:    true,
			Category:       foster_children.CategoryFatherless,
			BirthDate:      time.Date(2004, 6, 8, 0, 0, 0, 0, time.UTC),
			BirthPlace:     "Denpasar",
			SchoolName:     "SMAN 4 Denpasar",
			EducationLevel: 12,
			Address:        "Jl. Teuku Umar No. 88, Denpasar",
			FamilyCard:     "https://placehold.co/600x400.png",
			SKTM:           "https://placehold.co/600x400.png",
			CreatedAt:      now.Add(-time.Hour * 216),
			UpdatedAt:      now.Add(-time.Hour * 216),
		},
	}

	for _, child := range children {
		slug := fmt.Sprintf("%s-%s", pkg.Slugify(child.Name), child.ID.String()[:5])
		child.Slug = slug
		if err := db.Where("slug = ?", child.Slug).FirstOrCreate(&child).Error; err != nil {
			return fmt.Errorf("failed to seed foster child %s: %w", child.Name, err)
		}
	}

	fmt.Println("✅ Foster children seeded successfully!")
	return nil
}
