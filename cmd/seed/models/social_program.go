package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSocialPrograms(db *gorm.DB) error {
	fmt.Println("Seeding social programs...")

	now := time.Now()

	programs := []social_program.SocialProgram{
		{
			ID:            uuid.New(),
			Title:         "Orang Tua Asuh Penghafal Al-Qur'an",
			CoverImage:    "https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c?auto=format&fit=crop&w=800&q=80",
			Description:   "Program berkelanjutan untuk membiayai kebutuhan hidup dan pendidikan anak-anak penghafal Al-Qur'an di berbagai pondok pesantren dhuafa.",
			Status:        social_program.StatusActive,
			MinimumAmount: 100000,
			BillingDay:    5,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Beasiswa Pendidikan Anak Dhuafa",
			CoverImage:    "https://images.unsplash.com/photo-1503676260728-1c00da094a0b?auto=format&fit=crop&w=800&q=80",
			Description:   "Bantuan SPP bulanan dan perlengkapan sekolah untuk anak-anak dari keluarga prasejahtera agar mereka tetap dapat melanjutkan sekolah.",
			Status:        social_program.StatusActive,
			MinimumAmount: 50000,
			BillingDay:    10,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Orang Tua Asuh Anak Yatim",
			CoverImage:    "https://images.unsplash.com/photo-1488521787991-ed7bbaae773c?auto=format&fit=crop&w=800&q=80",
			Description:   "Program pengasuhan jarak jauh untuk memberikan kasih sayang berupa jaminan pemenuhan pangan, kesehatan, dan pendidikan bulanan anak yatim.",
			Status:        social_program.StatusActive,
			MinimumAmount: 150000,
			BillingDay:    5,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Dukungan Gizi Bayi & Balita Kurang Mampu",
			CoverImage:    "https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&w=800&q=80",
			Description:   "Program rutin penyediaan paket susu formula khusus, MPASI bergizi, dan vitamin bagi bayi serta balita dari keluarga prasejahtera untuk cegah stunting.",
			Status:        social_program.StatusActive,
			MinimumAmount: 75000,
			BillingDay:    15,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Kelas Kreatif Anak Jalanan",
			CoverImage:    "https://images.unsplash.com/photo-1516627145497-ae6968895b74?auto=format&fit=crop&w=800&q=80",
			Description:   "Penyelenggaraan kelas keterampilan mingguan (menggambar, musik, prakarya) dan penyediaan makan siang bergizi untuk anak-anak jalanan di kota besar.",
			Status:        social_program.StatusActive,
			MinimumAmount: 30000,
			BillingDay:    20,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Bantuan Operasional Sekolah Rumah Belajar",
			CoverImage:    "https://images.unsplash.com/photo-1427504494785-3a9ca7044f45?auto=format&fit=crop&w=800&q=80",
			Description:   "Program patungan rutin untuk biaya sewa tempat, listrik, internet, dan alat peraga edukatif pada sekolah non-formal gratis bagi anak marjinal.",
			Status:        social_program.StatusActive,
			MinimumAmount: 200000,
			BillingDay:    5,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Orang Tua Asuh Difabel Berprestasi",
			CoverImage:    "https://images.unsplash.com/photo-1531206715517-5c0ba140e2b8?auto=format&fit=crop&w=800&q=80",
			Description:   "Dukungan rutin bagi anak-anak berkebutuhan khusus yang berprestasi dalam bidang akademik, seni, maupun olahraga untuk mengembangkan bakat mereka.",
			Status:        social_program.StatusActive,
			MinimumAmount: 120000,
			BillingDay:    10,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Dukungan Pendidikan Anak Pesisir",
			CoverImage:    "https://images.unsplash.com/photo-1473116763269-255448993f66?auto=format&fit=crop&w=800&q=80",
			Description:   "Program pembiayaan transportasi perahu sekolah dan buku bacaan untuk anak-anak nelayan di pulau terluar agar mudah menjangkau sekolah.",
			Status:        social_program.StatusPending,
			MinimumAmount: 50000,
			BillingDay:    25,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New(),
			Title:         "Beasiswa Kuliah Calon Pemimpin Bangsa",
			CoverImage:    "https://images.unsplash.com/photo-1523050854058-8df90110c9f1?auto=format&fit=crop&w=800&q=80",
			Description:   "Program pembiayaan uang kuliah tunggal (UKT) bulanan bagi mahasiswa berprestasi nasional yang berasal dari latar belakang keluarga miskin.",
			Status:        social_program.StatusCompleted,
			MinimumAmount: 250000,
			BillingDay:    1,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:              uuid.New(),
			Title:           "Pembinaan Keterampilan Remaja Putus Sekolah",
			CoverImage:      "https://images.unsplash.com/photo-1521791136064-7986c2920216?auto=format&fit=crop&w=800&q=80",
			Description:     "Program bulanan untuk membiayai pelatihan menjahit, otomotif dasar, dan instalasi listrik bagi remaja putus sekolah agar siap bekerja.",
			Status:          social_program.StatusRejected,
			MinimumAmount:   80000,
			BillingDay:      15,
			RejectionReason: "Proposal detail tidak sesuai dengan format yang telah ditentukan yayasan",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	for i := range programs {
		programs[i].Slug = pkg.Slugify(programs[i].Title)

		var existing social_program.SocialProgram
		err := db.Where("slug = ?", programs[i].Slug).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&programs[i]).Error; err != nil {
					return fmt.Errorf("failed to create social program '%s': %w", programs[i].Title, err)
				}
				fmt.Printf("✓ Created social program: %s\n", programs[i].Title)
			} else {
				return fmt.Errorf("error checking existing social program: %w", err)
			}
		} else {
			fmt.Printf("⚠ Social program '%s' already exists, skipping...\n", programs[i].Title)
		}
	}

	return nil
}
