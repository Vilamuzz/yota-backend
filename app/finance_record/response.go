package finance_record

import "time"

type FinanceRecordResponse struct {
	ID              string    `json:"id"`
	FundType        string    `json:"fund_type"`
	FundID          string    `json:"fund_id"`
	SourceType      string    `json:"source_type"`
	SourceID        string    `json:"source_id"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toResponse(r *FinanceRecord) FinanceRecordResponse {
	return FinanceRecordResponse{
		ID:              r.ID,
		FundType:        r.FundType,
		FundID:          r.FundID,
		SourceType:      r.SourceType,
		SourceID:        r.SourceID,
		Amount:          r.Amount,
		TransactionDate: r.TransactionDate,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
