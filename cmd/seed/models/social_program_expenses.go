package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/social_program_expense"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSocialProgramExpenses(db *gorm.DB) error {
	fmt.Println("Seeding social program expenses...")

	var programs []social_program.SocialProgram
	if err := db.Find(&programs).Error; err != nil {
		return fmt.Errorf("failed to fetch social programs: %w", err)
	}

	// Find the finance account to attribute the expenses to
	var financeUser account.Account
	if err := db.Where("email = ?", "finance@yota.com").First(&financeUser).Error; err != nil {
		// Fallback to any user if finance is not found
		var fallbackUser account.Account
		if err := db.First(&fallbackUser).Error; err == nil {
			financeUser = fallbackUser
		} else {
			return fmt.Errorf("failed to find a user for CreatedBy field: %w", err)
		}
	}

	expenseTitles := []string{
		"Pembelian ATK dan Modul Belajar",
		"Sewa Ruang Belajar Tambahan",
		"Pembelian Seragam untuk Anak Yatim",
		"Penyediaan Susu Formula Balita",
		"Pembelian PMT (Makanan Tambahan)",
		"Honor Relawan Pengajar",
		"Konsumsi Kelas Parenting",
		"Pemeriksaan Kesehatan Anak Asuh",
		"Buku Dongeng & Alat Peraga Kreatif",
		"Laporan Evaluasi Program Sosial",
	}

	expenseNotes := []string{
		"Pembelian buku tulis, pensil warna, dan cetak modul pelajaran.",
		"Sewa ruang pertemuan warga untuk kelas akhir pekan.",
		"Seragam sekolah baru untuk anak-anak asuh yatim dhuafa.",
		"Pembelian 10 karton susu formula khusus balita stunting.",
		"Penyediaan bubur kacang hijau dan telur rebus untuk balita.",
		"Transportasi relawan pengajar kelas kreativitas anak.",
		"Makan siang peserta kelas pengasuhan balita sehat.",
		"Pemberian vitamin dan obat cacing rutin.",
		"Alat peraga edukasi untuk melatih motorik anak.",
		"Dokumentasi dan cetak laporan akhir kegiatan program sosial.",
	}

	expenseAmounts := []float64{
		15000, 25000, 35000, 20000, 40000, 30000, 15000, 20000, 25000, 10000,
	}

	now := time.Now()

	for _, program := range programs {
		fmt.Printf("Seeding expenses for social program: %s...\n", program.Title)
		for i := 0; i < 10; i++ {
			expenseID := uuid.New()
			expenseDate := now.AddDate(0, 0, -i)

			exp := social_program_expense.SocialProgramExpense{
				ID:              expenseID,
				SocialProgramID: program.ID,
				Title:           expenseTitles[i%len(expenseTitles)],
				Amount:          expenseAmounts[i%len(expenseAmounts)],
				ExpenseDate:     expenseDate,
				Note:            expenseNotes[i%len(expenseNotes)],
				ProofFile:       "https://placehold.co/600x400.png",
				CreatedBy:       financeUser.ID,
				CreatedAt:       expenseDate,
			}

			// Check if already exists by checking same program, title and date
			var existing social_program_expense.SocialProgramExpense
			err := db.Where("social_program_id = ? AND title = ? AND expense_date = ?", exp.SocialProgramID, exp.Title, exp.ExpenseDate).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&exp).Error; err != nil {
						return fmt.Errorf("failed to create social program expense %s: %w", exp.Title, err)
					}

					// Create a corresponding finance record for this expense (outflow)
					financeID := uuid.New().String()
					record := finance_record.FinanceRecord{
						ID:              financeID,
						FundType:        finance_record.FundTypeSocialProgram,
						FundID:          exp.SocialProgramID.String(),
						SourceType:      finance_record.SourceTypeExpense,
						SourceID:        exp.ID.String(),
						Amount:          exp.Amount,
						TransactionDate: exp.ExpenseDate,
						CreatedAt:       exp.CreatedAt,
					}
					if err := db.Create(&record).Error; err != nil {
						return fmt.Errorf("failed to create finance record for expense %s: %w", exp.Title, err)
					}
				} else {
					return fmt.Errorf("failed to check existing expense: %w", err)
				}
			}
		}
	}

	return nil
}