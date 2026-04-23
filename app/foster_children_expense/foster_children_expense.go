package foster_children_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/google/uuid"
)

type FosterChildrenExpense struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	FosterChildrenID uuid.UUID `json:"foster_children_id" gorm:"not null"`
	Title            string    `json:"title" gorm:"not null"`
	Amount           float64   `json:"amount" gorm:"not null"`
	ExpenseDate      time.Time `json:"expense_date" gorm:"not null"`
	Note             string    `json:"note" gorm:"not null"`
	ProofFile        string    `json:"proof_file"`
	CreatedBy        uuid.UUID `json:"created_by" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Account *account.Account `gorm:"foreignKey:CreatedBy;references:ID"`
}
