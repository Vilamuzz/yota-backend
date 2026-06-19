package models

import (
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/app/foster_children_expense"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedFosterChildrenExpenses(db *gorm.DB) error {
	fmt.Println("Seeding foster children expenses...")

	var children []foster_children.FosterChildren
	if err := db.Find(&children).Error; err != nil {
		return fmt.Errorf("failed to fetch foster children: %w", err)
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
		"Uang SPP Sekolah Bulanan",
		"Pembelian Buku Pelajaran Baru",
		"Pembelian Seragam Sekolah",
		"Sepatu dan Tas Sekolah Baru",
		"Alat Tulis dan Perlengkapan Belajar",
		"Uang Saku Harian",
		"Biaya Bimbingan Belajar",
		"Pemeriksaan Kesehatan Rutin",
		"Biaya Ekstrakurikuler Sekolah",
		"Kebutuhan Pribadi & Higienitas",
	}

	expenseNotes := []string{
		"Pembayaran SPP sekolah untuk bulan berjalan.",
		"Pembelian buku paket kurikulum merdeka.",
		"Seragam baru merah-putih/biru-putih lengkap dengan atribut.",
		"Sepatu sekolah hitam dan tas ransel baru.",
		"Buku tulis, pensil, pulpen, penggaris, dan kotak pensil.",
		"Dukungan uang saku harian selama 1 bulan.",
		"Les privat/bimbingan belajar mata pelajaran matematika.",
		"Check-up gigi dan kesehatan umum anak.",
		"Pembayaran kegiatan pramuka/olahraga.",
		"Pembelian perlengkapan mandi dan higienitas pribadi.",
	}

	expenseAmounts := []float64{
		150000, 75000, 200000, 150000, 50000, 100000, 120000, 80000, 60000, 50000,
	}

	now := time.Now()

	for _, child := range children {
		fmt.Printf("Seeding expenses for foster child: %s...\n", child.Name)
		for i := 0; i < 10; i++ {
			expenseID := uuid.New()
			expenseDate := now.AddDate(0, 0, -i)

			exp := foster_children_expense.FosterChildrenExpense{
				ID:               expenseID,
				FosterChildrenID: child.ID,
				Title:            expenseTitles[i%len(expenseTitles)],
				Amount:           expenseAmounts[i%len(expenseAmounts)],
				ExpenseDate:      expenseDate,
				Note:             expenseNotes[i%len(expenseNotes)],
				ProofFile:        "https://placehold.co/600x400.png",
				CreatedBy:        financeUser.ID,
				CreatedAt:        expenseDate,
				UpdatedAt:        expenseDate,
			}

			// Check if already exists by checking same child, title and date
			var existing foster_children_expense.FosterChildrenExpense
			err := db.Where("foster_children_id = ? AND title = ? AND expense_date = ?", exp.FosterChildrenID, exp.Title, exp.ExpenseDate).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&exp).Error; err != nil {
						return fmt.Errorf("failed to create foster child expense %s: %w", exp.Title, err)
					}

					// Create a corresponding finance record for this expense (outflow)
					financeID := uuid.New().String()
					record := finance_record.FinanceRecord{
						ID:              financeID,
						FundType:        finance_record.FundTypeFosterChildren,
						FundID:          exp.FosterChildrenID.String(),
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