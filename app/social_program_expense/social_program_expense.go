package social_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type SocialProgramExpense struct {
	ID              uuid.UUID `json:"id" gorm:"primaryKey"`
	SocialProgramID uuid.UUID `json:"social_program_id" gorm:"index;not null"`
	Title           string    `json:"title" gorm:"not null"`
	Amount          float64   `json:"amount" gorm:"not null"`
	ExpenseDate     time.Time `json:"expense_date" gorm:"not null"`
	Note            string    `json:"note" gorm:"not null"`
	ProofFile       string    `json:"proof_file"`
	CreatedBy       uuid.UUID `json:"created_by" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at"`

	Account *account.Account `gorm:"foreignKey:CreatedBy;references:ID"`
}
