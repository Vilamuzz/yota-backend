package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func makeSlug(title string) string {
	slug := strings.ToLower(title)
	// replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	// trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func SeedDonationPrograms(db *gorm.DB) error {
	fmt.Println("Seeding donation programs...")

	now := time.Now()

	programs := []donation_program.DonationProgram{
		{
			ID:          uuid.New(),
			Title:       "Bantuan Nutrisi untuk Balita Stunting",
			CoverImage:  "https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryHealth,
			Description: "Program bantuan penyediaan nutrisi tambahan, susu, dan vitamin untuk balita stunting di wilayah pedesaan terpencil demi mendukung tumbuh kembang anak secara maksimal.",
			FundTarget:  15000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, -1, 0),
			EndDate:     now.AddDate(0, 2, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Beasiswa Anak Yatim Piatu Berprestasi",
			CoverImage:  "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryEducation,
			Description: "Penyaluran beasiswa pendidikan untuk anak yatim piatu berprestasi tingkat SD hingga SMA agar mereka tetap dapat melanjutkan sekolah dan meraih cita-cita mereka.",
			FundTarget:  30000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, 0, -15),
			EndDate:     now.AddDate(0, 1, 15),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Pembangunan Sumur Air Bersih di Gunung Kidul",
			CoverImage:  "https://images.unsplash.com/photo-1541534741688-6078c6bfb5c5?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategorySocial,
			Description: "Penyediaan infrastruktur air bersih berupa pembuatan sumur bor dalam dan instalasi pipanisasi ke rumah-rumah warga yang terdampak kekeringan panjang.",
			FundTarget:  50000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, -2, 0),
			EndDate:     now.AddDate(0, 3, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Penanaman 10.000 Mangrove di Pantai Utara",
			CoverImage:  "https://images.unsplash.com/photo-1542601906990-b4d3fb778b09?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryEnvironment,
			Description: "Gerakan restorasi pesisir pantai utara Jawa melalui penanaman bibit pohon mangrove untuk mencegah abrasi, melindungi habitat pesisir, dan mengembalikan ekosistem laut.",
			FundTarget:  20000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, 0, -5),
			EndDate:     now.AddDate(0, 1, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Bantuan Darurat Korban Gempa Bumi",
			CoverImage:  "https://images.unsplash.com/photo-1488521787991-ed7bbaae773c?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryDisaster,
			Description: "Tanggap darurat bencana untuk penyaluran logistik, tenda darurat, dapur umum, obat-obatan, dan selimut bagi korban gempa bumi yang kehilangan tempat tinggal.",
			FundTarget:  100000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, 0, -2),
			EndDate:     now.AddDate(0, 0, 28),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Pondok Lansia Sehat & Mandiri",
			CoverImage:  "https://images.unsplash.com/photo-1508847154043-be12a26c86c1?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryHumanity,
			Description: "Dukungan operasional dan penyediaan layanan kesehatan gratis serta pemenuhan pangan bergizi sehari-hari untuk lansia terlantar agar mereka dapat hidup layak.",
			FundTarget:  40000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, -1, -15),
			EndDate:     now.AddDate(0, 2, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Perbaikan Jembatan Desa Tertinggal",
			CoverImage:  "https://images.unsplash.com/photo-1517048676732-d65bc937f952?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryOther,
			Description: "Renovasi dan pembangunan jembatan penyeberangan sungai desa yang rusak parah agar akses transportasi warga, petani, dan anak sekolah kembali aman.",
			FundTarget:  45000000,
			Status:      donation_program.StatusActive,
			StartDate:   now.AddDate(0, 0, -10),
			EndDate:     now.AddDate(0, 2, 10),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Renovasi Perpustakaan Sekolah Terpencil",
			CoverImage:  "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryEducation,
			Description: "Merenovasi gedung perpustakaan sekolah yang bocor dan melengkapinya dengan ratusan buku bacaan baru serta meja belajar yang layak bagi siswa.",
			FundTarget:  25000000,
			Status:      donation_program.StatusDraft,
			StartDate:   now.AddDate(0, 1, 0),
			EndDate:     now.AddDate(0, 3, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Operasi Katarak Gratis untuk Lansia Kurang Mampu",
			CoverImage:  "https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategoryHealth,
			Description: "Penyelenggaraan operasi katarak gratis bagi lansia yang memiliki gangguan penglihatan namun memiliki keterbatasan ekonomi untuk berobat.",
			FundTarget:  60000000,
			Status:      donation_program.StatusCompleted,
			StartDate:   now.AddDate(0, -3, 0),
			EndDate:     now.AddDate(0, -1, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Title:       "Bantuan Sembako untuk Pekerja Harian",
			CoverImage:  "https://images.unsplash.com/photo-1541534741688-6078c6bfb5c5?auto=format&fit=crop&w=800&q=80",
			Category:    donation_program.CategorySocial,
			Description: "Penyaluran paket sembako gratis untuk meringankan beban ekonomi keluarga pekerja harian lepas, buruh cuci, dan buruh tani.",
			FundTarget:  15000000,
			Status:      donation_program.StatusExpired,
			StartDate:   now.AddDate(0, -2, 0),
			EndDate:     now.AddDate(0, -1, 0),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for i := range programs {
		programs[i].Slug = makeSlug(programs[i].Title)

		var existing donation_program.DonationProgram
		err := db.Where("slug = ?", programs[i].Slug).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&programs[i]).Error; err != nil {
					return fmt.Errorf("failed to create donation program '%s': %w", programs[i].Title, err)
				}
				fmt.Printf("✓ Created donation program: %s\n", programs[i].Title)
			} else {
				return fmt.Errorf("error checking existing donation program: %w", err)
			}
		} else {
			fmt.Printf("⚠ Donation program '%s' already exists, skipping...\n", programs[i].Title)
		}
	}

	return nil
}