package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/app/ambulance_service_request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedAmbulanceServiceRequests(db *gorm.DB) error {
	fmt.Println("Seeding ambulance service requests...")

	var ambulances []ambulance.Ambulance
	if err := db.Find(&ambulances).Error; err != nil {
		return fmt.Errorf("failed to fetch ambulances: %w", err)
	}

	if len(ambulances) == 0 {
		return fmt.Errorf("no ambulances found to link requests to")
	}

	var users []account.Account
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found to submit requests")
	}

	patientNames := []string{
		"Joko Susilo", "Sri Wahyuni", "Hendra Wijaya", "Kartika Sari",
		"Bambang Pamungkas", "Endang Rahayu", "Rudi Hermawan", "Ani Lestari",
		"Taufik Hidayat", "Siti Rahma",
	}

	patientAddresses := []string{
		"Jl. Kenanga No. 12, RT 03/RW 04, Jakarta",
		"Jl. Mawar No. 45, RT 01/RW 02, Bandung",
		"Jl. Melati No. 8, RT 05/RW 01, Surabaya",
		"Jl. Anggrek No. 19, RT 02/RW 03, Yogyakarta",
		"Jl. Dahlia No. 30, RT 04/RW 05, Semarang",
		"Jl. Kamboja No. 14, RT 03/RW 01, Malang",
		"Jl. Flamboyan No. 7, RT 01/RW 04, Solo",
		"Jl. Tulip No. 22, RT 02/RW 02, Cirebon",
		"Jl. Sakura No. 11, RT 05/RW 03, Bogor",
		"Jl. Lotus No. 5, RT 04/RW 01, Tangerang",
	}

	destinations := []string{
		"RS Pusat Angkatan Darat (RSPAD) Gatot Soebroto",
		"RSUP rujukan nasional Dr. Cipto Mangunkusumo",
		"RS Harapan Kita",
		"RS Fatmawati",
		"RSPON Jakarta",
		"RS Siloam Semanggi",
		"RS Medistra",
		"RS Pondok Indah",
		"RS Pertamina Pusat",
		"RS Hermina",
	}

	diseases := []string{
		"Stroke ringan, perlu penanganan segera",
		"Demam tinggi dan sesak napas",
		"Patah tulang kaki kanan pasca jatuh",
		"Jadwal rutin cuci darah (hemodialisis)",
		"Kontrol pasca operasi jantung",
		"Pasien kritis ICU transfer rumah sakit",
		"Persalinan darurat pembukaan 8",
		"Asma akut kambuh",
		"Pasien pasca stroke kontrol syaraf",
		"Kecelakaan ringan di rumah",
	}

	categories := []ambulance_history.ServiceCategory{
		ambulance_history.EmergencyService,
		ambulance_history.PatientService,
		ambulance_history.EmergencyService,
		ambulance_history.SocialService,
		ambulance_history.PatientService,
		ambulance_history.EmergencyService,
		ambulance_history.MortuaryService,
		ambulance_history.EmergencyService,
		ambulance_history.PatientService,
		ambulance_history.OtherService,
	}

	now := time.Now()

	for _, amb := range ambulances {
		fmt.Printf("Seeding service requests for ambulance plate: %s...\n", amb.PlateNumber)
		for i := 0; i < 10; i++ {
			reqID := uuid.New()
			// Spread pick up times over the past 10 days
			pickupDate := now.AddDate(0, 0, -i)

			var status ambulance_service_request.Status
			switch i {
			case 7:
				status = ambulance_service_request.StatusInService
			case 8:
				status = ambulance_service_request.StatusAccepted
			case 9:
				status = ambulance_service_request.StatusPending
			default:
				status = ambulance_service_request.StatusDone
			}

			// We assign the submitter user dynamically from the available users
			submitter := users[i%len(users)]

			ambID := amb.ID
			req := ambulance_service_request.AmbulanceServiceRequest{
				ID:              reqID,
				SubmittedBy:     submitter.ID,
				AmbulanceID:     &ambID,
				SubmitterName:   submitter.Email, // default to submitter email or fallback
				SubmitterPhone:  "081234567890",
				SubmitterIDCard: "https://placehold.co/600x400.png",
				PatientName:     patientNames[i%len(patientNames)],
				PatientAddress:  patientAddresses[i%len(patientAddresses)],
				PatientAge:      20 + (i * 5),
				IsInfectious:    i%5 == 0,
				Disease:         diseases[i%len(diseases)],
				IsAbleToSit:     i%2 == 0,
				PickupDate:      pickupDate,
				PickupTime:      pickupDate,
				Destination:     destinations[i%len(destinations)],
				Note:            "Penanganan darurat mohon dibantu secepatnya.",
				Status:          status,
				ServiceCategory: categories[i%len(categories)],
				CreatedAt:       pickupDate.Add(-time.Hour),
				UpdatedAt:       pickupDate,
			}

			// Check if already exists
			var existing ambulance_service_request.AmbulanceServiceRequest
			err := db.Where("ambulance_id = ? AND patient_name = ? AND status = ?", req.AmbulanceID, req.PatientName, req.Status).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&req).Error; err != nil {
						return fmt.Errorf("failed to create ambulance service request: %w", err)
					}

					// If Status is Done, also create an AmbulanceHistory record
					if status == ambulance_service_request.StatusDone {
						note := fmt.Sprintf("Layanan ambulans selesai untuk permintaan dari %s. Pasien: %s", req.SubmitterName, req.PatientName)
						history := ambulance_history.AmbulanceHistory{
							ID:              uuid.New(),
							AmbulanceID:     amb.ID,
							DriverID:        amb.DriverID,
							ServiceCategory: req.ServiceCategory,
							Note:            note,
							CreatedAt:       req.UpdatedAt,
						}
						if err := db.Create(&history).Error; err != nil {
							return fmt.Errorf("failed to create ambulance history record for request %s: %w", req.ID.String(), err)
						}
					}
				} else {
					return fmt.Errorf("failed to check existing request: %w", err)
				}
			}
		}
	}

	return nil
}