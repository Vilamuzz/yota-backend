package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/foster_children_candidate"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedFosterChildrenCandidates(db *gorm.DB) error {
	fmt.Println("Seeding foster children candidates...")

	var users []account.Account
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found to submit candidates")
	}

	names := []string{
		"Andi Wijaya", "Rina Lestari", "Tono Hartono", "Siti Rahayu",
		"Dani Saputra", "Mega Utami", "Feri Irawan", "Lilis Suryani",
		"Hadi Pranoto", "Yanti Rosmiati",
	}

	niks := []string{
		"3172012501120001",
		"3273082508140002",
		"3578110209100003",
		"3374011712130004",
		"3404303006110005",
		"3573121504090006",
		"1271070718080007",
		"1371012211070008",
		"1671101003120009",
		"5171080619110010",
	}

	profilePictures := []string{
		"https://images.unsplash.com/photo-1580489944761-15a19d654956?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1517841905240-472988babdf9?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1531123897727-8f129e1688ce?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1544725176-7c40e5a71c5e?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1522075469751-3a6694fb2f61?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1503919545889-aef636e10ad4?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=600&auto=format&fit=crop&q=80",
		"https://images.unsplash.com/photo-1595152772835-219674b2a8a6?w=600&auto=format&fit=crop&q=80",
	}

	genders := []foster_children_candidate.Gender{
		foster_children_candidate.Male,
		foster_children_candidate.Female,
		foster_children_candidate.Male,
		foster_children_candidate.Female,
		foster_children_candidate.Male,
		foster_children_candidate.Female,
		foster_children_candidate.Male,
		foster_children_candidate.Female,
		foster_children_candidate.Male,
		foster_children_candidate.Female,
	}

	categories := []foster_children_candidate.Category{
		foster_children_candidate.CategoryFatherless,
		foster_children_candidate.CategoryMotherless,
		foster_children_candidate.CategoryOrphan,
		foster_children_candidate.CategoryFatherless,
		foster_children_candidate.CategoryMotherless,
		foster_children_candidate.CategoryOrphan,
		foster_children_candidate.CategoryFatherless,
		foster_children_candidate.CategoryMotherless,
		foster_children_candidate.CategoryOrphan,
		foster_children_candidate.CategoryFatherless,
	}

	birthPlaces := []string{
		"Jakarta", "Bandung", "Surabaya", "Yogyakarta", "Semarang",
		"Malang", "Solo", "Cirebon", "Bogor", "Tangerang",
	}

	schools := []string{
		"SDN Menteng 01", "SMPN 2 Bandung", "SDN Bubutan", "SMPN 1 Yogyakarta", "SDN Gajahmungkur",
		"SMPN 4 Malang", "SDN Pajang", "SMPN 3 Cirebon", "SDN Baranangsiang", "SMPN 1 Tangerang",
	}

	addresses := []string{
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

	now := time.Now()

	for i := 0; i < 10; i++ {
		candidateID := uuid.New()
		submitter := users[i%len(users)]

		var status foster_children_candidate.Status
		var rejectionReason string

		switch i {
		case 5:
			status = foster_children_candidate.StatusSocialManagerAccepted
		case 6:
			status = foster_children_candidate.StatusSocialManagerAccepted
		case 7:
			status = foster_children_candidate.StatusAccepted
		case 8:
			status = foster_children_candidate.StatusRejected
			rejectionReason = "Dokumen persyaratan kurang lengkap (SKTM tidak terbaca)"
		case 9:
			status = foster_children_candidate.StatusCancelled
		default:
			status = foster_children_candidate.StatusPending
		}

		cand := foster_children_candidate.FosterChildrenCandidate{
			ID:               candidateID,
			Name:             names[i],
			Nik:              niks[i],
			ProfilePicture:   profilePictures[i],
			Gender:           genders[i],
			Category:         categories[i],
			BirthDate:        time.Date(2012+i/3, time.Month(i+1), 5+i*2, 0, 0, 0, 0, time.UTC),
			BirthPlace:       birthPlaces[i],
			SchoolName:       schools[i],
			EducationLevel:   2 + i,
			Address:          addresses[i],
			FamilyCard:       "https://placehold.co/600x400.png",
			SKTM:             "https://placehold.co/600x400.png",
			SubmitterName:    submitter.Email,
			SubmitterPhone:   "081234567890",
			SubmitterAddress: "Jl. Contoh Submitter No. " + fmt.Sprintf("%d", i+1),
			SubmitterIDCard:  "https://placehold.co/600x400.png",
			SubmittedBy:      submitter.ID,
			Status:           status,
			RejectionReason:  rejectionReason,
			CreatedAt:        now.AddDate(0, 0, -i),
			UpdatedAt:        now.AddDate(0, 0, -i),
		}

		// Check if already exists by checking name and birthDate
		var existing foster_children_candidate.FosterChildrenCandidate
		err := db.Where("name = ? AND birth_date = ?", cand.Name, cand.BirthDate).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&cand).Error; err != nil {
					return fmt.Errorf("failed to create foster child candidate %s: %w", cand.Name, err)
				}
				fmt.Printf("✓ Created foster child candidate: %s\n", cand.Name)
			} else {
				return fmt.Errorf("failed to check existing candidate: %w", err)
			}
		}
	}

	return nil
}
