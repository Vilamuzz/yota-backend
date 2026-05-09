package finance_record

import "time"

// FundType identifies what fund the record belongs to
const (
	FundTypeDonation       = "donation"
	FundTypeFosterChildren = "foster_children"
	FundTypeSocialProgram  = "social_program"
)

// SourceType identifies what triggered the record
// transaction = income
// expense = outflow
const (
	SourceTypeTransaction = "transaction"
	SourceTypeExpense     = "expense"
)

type FinanceRecord struct {
	ID              string     `json:"id" gorm:"primaryKey"`
	FundType        string     `json:"fundType"`
	FundID          string     `json:"fundId"`
	SourceType      string     `json:"sourceType"`
	SourceID        string     `json:"sourceId"`
	Amount          float64    `json:"amount"`
	TransactionDate time.Time  `json:"transactionDate"`
	CreatedAt       time.Time  `json:"createdAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}
