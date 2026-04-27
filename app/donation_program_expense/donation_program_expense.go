package donation_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type DonationProgramExpense struct {
	ID                uuid.UUID `json:"id" gorm:"primaryKey"`
	DonationProgramID uuid.UUID `json:"donationProgramId" gorm:"index;not null"`
	Title             string    `json:"title" gorm:"not null"`
	Amount            float64   `json:"amount" gorm:"not null"`
	ExpenseDate       time.Time `json:"expenseDate" gorm:"not null"`
	Note              string    `json:"note" gorm:"not null"`
	ProofFile         string    `json:"proofFile"`
	CreatedBy         uuid.UUID `json:"createdBy" gorm:"not null"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`

	Account account.Account `json:"-" gorm:"foreignKey:CreatedBy;references:ID"`
}
