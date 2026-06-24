package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/donation_program_expense"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedDonationExpenses(db *gorm.DB) error {
	fmt.Println("Seeding donation expenses...")

	var programs []donation_program.DonationProgram
	if err := db.Find(&programs).Error; err != nil {
		return fmt.Errorf("failed to fetch donation programs: %w", err)
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
		"Pembelian Sembako Tahap 1",
		"Biaya Operasional Penyaluran",
		"Pembelian Vitamin dan Susu Formula",
		"Sewa Transportasi Logistik",
		"Konsumsi Relawan Lapangan",
		"Pembelian Sembako Tahap 2",
		"Biaya Administrasi Bank",
		"Alat Tulis Kantor dan Dokumentasi",
		"Pembelian Masker dan Sanitizer",
		"Evaluasi dan Pelaporan Program",
	}

	expenseNotes := []string{
		"Pembelian sembako untuk didistribusikan ke penerima manfaat.",
		"Biaya BBM dan tol untuk pengiriman bantuan.",
		"Untuk menunjang pemenuhan gizi anak-anak.",
		"Sewa mobil box selama 2 hari untuk distribusi.",
		"Makan siang relawan selama kegiatan berlangsung.",
		"Pembelian sembako tambahan karena kuota bertambah.",
		"Biaya transfer antar bank dan admin transaksi.",
		"Cetak proposal dan laporan pertanggungjawaban.",
		"Protokol kesehatan selama penyaluran bantuan.",
		"Penyusunan laporan akhir program donasi.",
	}

	expenseAmounts := []float64{
		50000, 100000, 150000, 75000, 200000, 120000, 25000, 80000, 60000, 150000,
	}

	now := time.Now()

	for _, program := range programs {
		fmt.Printf("Seeding expenses for program: %s...\n", program.Title)
		for i := 0; i < 10; i++ {
			expenseID := uuid.New()
			expenseDate := now.AddDate(0, 0, -i)

			exp := donation_program_expense.DonationProgramExpense{
				ID:                expenseID,
				DonationProgramID: program.ID,
				Title:             expenseTitles[i%len(expenseTitles)],
				Amount:            expenseAmounts[i%len(expenseAmounts)],
				ExpenseDate:       expenseDate,
				Note:              expenseNotes[i%len(expenseNotes)],
				ProofFile:         "https://placehold.co/600x400.png",
				CreatedBy:         financeUser.ID,
				CreatedAt:         expenseDate,
				UpdatedAt:         expenseDate,
			}

			// Check if already exists by checking same program, title and date
			var existing donation_program_expense.DonationProgramExpense
			err := db.Where("donation_program_id = ? AND title = ? AND expense_date = ?", exp.DonationProgramID, exp.Title, exp.ExpenseDate).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&exp).Error; err != nil {
						return fmt.Errorf("failed to create donation program expense %s: %w", exp.Title, err)
					}

					// Create a corresponding finance record for this expense (outflow)
					financeID := uuid.New().String()
					record := finance_record.FinanceRecord{
						ID:              financeID,
						FundType:        finance_record.FundTypeDonation,
						FundID:          exp.DonationProgramID.String(),
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