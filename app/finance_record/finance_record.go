package finance_record

import "time"

// FundType identifies what fund the record belongs to
const (
	FundTypeDonation = "donation"
)

// SourceType identifies what triggered the record
// transaction = income
// expense = outflow
const (
	SourceTypeTransaction = "transaction"
	SourceTypeExpense     = "expense"
)

type FinanceRecord struct {
	ID              string    `json:"id"               gorm:"primaryKey"`
	FundType        string    `json:"fund_type"`
	FundID          string    `json:"fund_id"`
	SourceType      string    `json:"source_type"`
	SourceID        string    `json:"source_id"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
