package finance_record

import "time"

type FinanceRecordResponse struct {
	ID              string    `json:"id"`
	FundType        string    `json:"fundType"`
	FundID          string    `json:"fundId"`
	SourceType      string    `json:"sourceType"`
	SourceID        string    `json:"sourceId"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transactionDate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
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
	}
}
