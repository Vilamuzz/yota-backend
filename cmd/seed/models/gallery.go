package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/gallery"
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func makeGallerySlug(title string) string {
	slug := strings.ToLower(title)
	// replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	// trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func SeedGallery(db *gorm.DB) error {
	fmt.Println("Seeding gallery...")

	now := time.Now()

	galleries := []gallery.Gallery{
		{
			ID:          uuid.New(),
			Title:       "Aktivitas Pembagian Nutrisi untuk Anak",
			Category:    media.Health,
			CoverImage:  "https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&w=800&q=80",
			Status:      media.MediaStatusPublished,
			Description: "Dokumentasi foto kegiatan pembagian makanan tambahan bergizi dan susu kepada balita stunting di desa binaan.",
			Views:       45,
			CreatedAt:   now.AddDate(0, 0, -10),
			UpdatedAt:   now.AddDate(0, 0, -10),
		},
		{
			ID:          uuid.New(),
			Title:       "Penyaluran Sembako Darurat Bencana",
			Category:    media.Disaster,
			CoverImage:  "https://images.unsplash.com/photo-1488521787991-ed7bbaae773c?auto=format&fit=crop&w=800&q=80",
			Status:      media.MediaStatusPublished,
			Description: "Kumpulan foto proses distribusi sembako, selimut, dan tenda darurat oleh relawan Yota di lokasi bencana gempa bumi.",
			Views:       120,
			CreatedAt:   now.AddDate(0, 0, -2),
			UpdatedAt:   now.AddDate(0, 0, -2),
		},
		{
			ID:          uuid.New(),
			Title:       "Aksi Bersih Pantai & Tanam Mangrove",
			Category:    media.Environment,
			CoverImage:  "https://images.unsplash.com/photo-1542601906990-b4d3fb778b09?auto=format&fit=crop&w=800&q=80",
			Status:      media.MediaStatusPublished,
			Description: "Dokumentasi foto kolaborasi warga sekitar dan relawan dalam menjaga kebersihan lingkungan pantai dan restorasi mangrove.",
			Views:       67,
			CreatedAt:   now.AddDate(0, 0, -5),
			UpdatedAt:   now.AddDate(0, 0, -5),
		},
		{
			ID:          uuid.New(),
			Title:       "Kunjungan Kasih ke Panti Jompo Sejahtera",
			Category:    media.SocialEvent,
			CoverImage:  "https://images.unsplash.com/photo-1508847154043-be12a26c86c1?auto=format&fit=crop&w=800&q=80",
			Status:      media.MediaStatusPublished,
			Description: "Momen kebersamaan relawan Yota dengan para lansia, melakukan cek kesehatan gratis, serta berbagi bingkisan kebahagiaan.",
			Views:       88,
			CreatedAt:   now.AddDate(0, 0, -15),
			UpdatedAt:   now.AddDate(0, 0, -15),
		},
		{
			ID:          uuid.New(),
			Title:       "Draft Galeri Foto Penyerahan Beasiswa",
			Category:    media.SocialEvent,
			CoverImage:  "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?auto=format&fit=crop&w=800&q=80",
			Status:      media.MediaStatusDraft,
			Description: "Draf dokumentasi penyerahan bantuan beasiswa pendidikan Yota untuk tahun ajaran baru.",
			Views:       0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for i := range galleries {
		galleries[i].Slug = makeGallerySlug(galleries[i].Title)

		// Create associated media items for the gallery
		galleryID := galleries[i].ID
		galleries[i].Media = []media.Media{
			{
				ID:        uuid.New(),
				GalleryID: &galleryID,
				Type:      media.MediaTypeImage,
				URL:       galleries[i].CoverImage,
				Alt:       galleries[i].Title,
				Order:     0,
				CreatedAt: galleries[i].CreatedAt,
				UpdatedAt: galleries[i].UpdatedAt,
			},
			{
				ID:        uuid.New(),
				GalleryID: &galleryID,
				Type:      media.MediaTypeImage,
				URL:       "https://images.unsplash.com/photo-1488521787991-ed7bbaae773c?auto=format&fit=crop&w=800&q=80",
				Alt:       fmt.Sprintf("%s - Detail 1", galleries[i].Title),
				Order:     1,
				CreatedAt: galleries[i].CreatedAt,
				UpdatedAt: galleries[i].UpdatedAt,
			},
		}

		var existing gallery.Gallery
		err := db.Where("slug = ?", galleries[i].Slug).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&galleries[i]).Error; err != nil {
					return fmt.Errorf("failed to create gallery item '%s': %w", galleries[i].Title, err)
				}
				fmt.Printf("✓ Created gallery: %s\n", galleries[i].Title)
			} else {
				return fmt.Errorf("error checking existing gallery: %w", err)
			}
		} else {
			fmt.Printf("⚠ Gallery '%s' already exists, skipping...\n", galleries[i].Title)
		}
	}

	return nil
}
