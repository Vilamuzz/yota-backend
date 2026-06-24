package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedAmbulanceHistories(db *gorm.DB) error {
	fmt.Println("Seeding ambulance histories...")

	var ambulances []ambulance.Ambulance
	if err := db.Find(&ambulances).Error; err != nil {
		return fmt.Errorf("failed to fetch ambulances: %w", err)
	}

	notes := []string{
		"Rujukan pasien darurat stroke dari Puskesmas ke RSUD.",
		"Layanan pengantaran jenazah ke TPU Pondok Ranggon.",
		"Penjemputan pasien kontrol pasca operasi bypass jantung.",
		"Layanan siaga ambulance untuk kegiatan bakti sosial warga.",
		"Rujukan pasien ibu melahirkan dengan penyulit (preeklampsia).",
		"Pengantaran pasien cuci darah (hemodialisis) rutin.",
		"Transfer pasien ICU antar rumah sakit dengan ventilator.",
		"Pertolongan pertama korban kecelakaan lalu lintas jalan raya.",
		"Pengantaran jenazah ke luar kota (Bogor).",
		"Layanan antar jemput pasien lansia untuk kontrol kesehatan rutin.",
	}

	categories := []ambulance_history.ServiceCategory{
		ambulance_history.EmergencyService,
		ambulance_history.MortuaryService,
		ambulance_history.PatientService,
		ambulance_history.SocialService,
		ambulance_history.EmergencyService,
		ambulance_history.PatientService,
		ambulance_history.EmergencyService,
		ambulance_history.EmergencyService,
		ambulance_history.MortuaryService,
		ambulance_history.PatientService,
	}

	now := time.Now()

	for _, amb := range ambulances {
		fmt.Printf("Seeding 10 histories for ambulance plate: %s...\n", amb.PlateNumber)
		for i := 0; i < 10; i++ {
			createdAt := now.AddDate(0, 0, -i)
			hist := ambulance_history.AmbulanceHistory{
				ID:              uuid.New(),
				AmbulanceID:     amb.ID,
				DriverID:        amb.DriverID,
				ServiceCategory: categories[i%len(categories)],
				Note:            notes[i%len(notes)],
				CreatedAt:       createdAt,
			}

			// Check if already exists for this ambulance, note and date
			var existing ambulance_history.AmbulanceHistory
			err := db.Where("ambulance_id = ? AND note = ? AND created_at = ?", hist.AmbulanceID, hist.Note, hist.CreatedAt).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&hist).Error; err != nil {
						return fmt.Errorf("failed to create ambulance history: %w", err)
					}
				} else {
					return fmt.Errorf("failed to check existing history: %w", err)
				}
			}
		}
	}

	return nil
}