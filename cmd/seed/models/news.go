package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func makeNewsSlug(title string) string {
	slug := strings.ToLower(title)
	// replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	// trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func SeedNews(db *gorm.DB) error {
	fmt.Println("Seeding news...")

	now := time.Now()
	publishedAt1 := now.AddDate(0, 0, -10)
	publishedAt2 := now.AddDate(0, 0, -2)
	publishedAt3 := now.AddDate(0, 0, -5)
	publishedAt4 := now.AddDate(0, 0, -15)

	articles := []news.News{
		{
			ID:         uuid.New(),
			Title:      "Yota Salurkan Bantuan Nutrisi untuk Balita Stunting di Gunungkidul",
			CoverImage: "https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&w=800&q=80",
			Category:   media.Health,
			Content:    "<p>Yayasan Yota telah menyalurkan bantuan nutrisi tambahan untuk puluhan balita stunting di Gunungkidul. Program ini bertujuan membantu tumbuh kembang anak secara optimal dengan memberikan susu formula, vitamin, serta makanan tambahan padat nutrisi.</p><p>Kegiatan ini merupakan bagian dari komitmen Yota untuk menekan angka stunting di daerah terpencil.</p>",
			Status:     media.MediaStatusPublished,
			Views:      152,
			PublishedAt: &publishedAt1,
			CreatedAt:  now.AddDate(0, 0, -10),
			UpdatedAt:  now.AddDate(0, 0, -10),
		},
		{
			ID:         uuid.New(),
			Title:      "Relawan Yota Bantu Evakuasi dan Logistik Korban Gempa Bumi",
			CoverImage: "https://images.unsplash.com/photo-1488521787991-ed7bbaae773c?auto=format&fit=crop&w=800&q=80",
			Category:   media.Disaster,
			Content:    "<p>Tim relawan tanggap bencana Yota segera dikerahkan ke lokasi gempa untuk membantu evakuasi warga dan mendistribusikan bantuan logistik darurat. Bantuan yang disalurkan meliputi tenda darurat, selimut, makanan instan, dan obat-obatan dasar.</p><p>Yota terus berkoordinasi dengan pihak berwenang setempat untuk memastikan bantuan tersalurkan dengan merata.</p>",
			Status:     media.MediaStatusPublished,
			Views:      340,
			PublishedAt: &publishedAt2,
			CreatedAt:  now.AddDate(0, 0, -2),
			UpdatedAt:  now.AddDate(0, 0, -2),
		},
		{
			ID:         uuid.New(),
			Title:      "Aksi Tanam 10.000 Mangrove Yota di Pesisir Utara Jawa",
			CoverImage: "https://images.unsplash.com/photo-1542601906990-b4d3fb778b09?auto=format&fit=crop&w=800&q=80",
			Category:   media.Environment,
			Content:    "<p>Sebagai bagian dari upaya pelestarian lingkungan dan pencegahan abrasi pantai, Yota bersama komunitas lokal menyelenggarakan aksi penanaman 10.000 bibit mangrove di pesisir utara Jawa.</p><p>Langkah nyata ini diharapkan dapat menjaga ekosistem pesisir serta memberikan dampak positif jangka panjang bagi nelayan sekitar.</p>",
			Status:     media.MediaStatusPublished,
			Views:      89,
			PublishedAt: &publishedAt3,
			CreatedAt:  now.AddDate(0, 0, -5),
			UpdatedAt:  now.AddDate(0, 0, -5),
		},
		{
			ID:         uuid.New(),
			Title:      "Yota Membuka Program Beasiswa untuk Anak Yatim Piatu",
			CoverImage: "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?auto=format&fit=crop&w=800&q=80",
			Category:   media.SocialEvent,
			Content:    "<p>Yayasan Yota resmi membuka pendaftaran program beasiswa khusus untuk anak yatim piatu berprestasi. Program ini mencakup bantuan biaya sekolah, perlengkapan belajar, hingga pendampingan karakter bagi para penerima manfaat.</p><p>Diharapkan program ini dapat meringankan beban pendidikan mereka dan menyalakan harapan masa depan.</p>",
			Status:     media.MediaStatusPublished,
			Views:      210,
			PublishedAt: &publishedAt4,
			CreatedAt:  now.AddDate(0, 0, -15),
			UpdatedAt:  now.AddDate(0, 0, -15),
		},
		{
			ID:         uuid.New(),
			Title:      "Rencana Program Pembangunan Sumur Air Bersih Berikutnya",
			CoverImage: "https://images.unsplash.com/photo-1541534741688-6078c6bfb5c5?auto=format&fit=crop&w=800&q=80",
			Category:   media.Others,
			Content:    "<p>Ini adalah draf rencana kegiatan pembangunan sumur air bersih berikutnya untuk daerah yang mengalami kekeringan ekstrem. Tim kami sedang melakukan survei lapangan untuk menentukan titik lokasi pengeboran terbaik.</p>",
			Status:     media.MediaStatusDraft,
			Views:      0,
			PublishedAt: nil,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	for i := range articles {
		articles[i].Slug = makeNewsSlug(articles[i].Title)

		// Include some default media items for the news (simulating gallery inside news)
		newsID := articles[i].ID
		articles[i].Media = []media.Media{
			{
				ID:        uuid.New(),
				NewsID:    &newsID,
				Type:      media.MediaTypeImage,
				URL:       articles[i].CoverImage,
				Alt:       articles[i].Title,
				Order:     0,
				CreatedAt: articles[i].CreatedAt,
				UpdatedAt: articles[i].UpdatedAt,
			},
		}

		var existing news.News
		err := db.Where("slug = ?", articles[i].Slug).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&articles[i]).Error; err != nil {
					return fmt.Errorf("failed to create news article '%s': %w", articles[i].Title, err)
				}
				fmt.Printf("✓ Created news article: %s\n", articles[i].Title)
			} else {
				return fmt.Errorf("error checking existing news article: %w", err)
			}
		} else {
			fmt.Printf("⚠ News article '%s' already exists, skipping...\n", articles[i].Title)
		}
	}

	return nil
}
