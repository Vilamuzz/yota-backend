package foster_child_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildTransactionResponse struct {
	ID                int        `json:"id"`
	FosterChildID     int        `json:"foster_child_id"`
	OrderID           string     `json:"order_id"`
	DonorName         string     `json:"donor_name"`
	DonorEmail        string     `json:"donor_email"`
	Source            bool       `json:"source"`
	GrossAmount       float64    `json:"gross_amount"`
	FraudStatus       string     `json:"fraud_status"`
	TransactionStatus string     `json:"transaction_status"`
	Provider          string     `json:"provider"`
	TransactionID     string     `json:"transaction_id"`
	SnapToken         string     `json:"snap_token"`
	SnapRedirectURL   string     `json:"snap_redirect_url"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type FosterChildTransactionListResponse struct {
	Transactions []FosterChildTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination             `json:"pagination"`
}

func (tx *FosterChildTransaction) toFosterChildTransactionResponse() FosterChildTransactionResponse {
	return FosterChildTransactionResponse{
		ID:                tx.ID,
		FosterChildID:     tx.FosterChildID,
		OrderID:           tx.OrderID,
		DonorName:         tx.DonorName,
		DonorEmail:        tx.DonorEmail,
		Source:            tx.Source,
		GrossAmount:       tx.GrossAmount,
		FraudStatus:       tx.FraudStatus,
		TransactionStatus: tx.TransactionStatus,
		Provider:          tx.Provider,
		TransactionID:     tx.TransactionID,
		SnapToken:         tx.SnapToken,
		SnapRedirectURL:   tx.SnapRedirectURL,
		PaidAt:            tx.PaidAt,
		CreatedAt:         tx.CreatedAt,
	}
}

func toFosterChildTransactionListResponse(transactions []FosterChildTransaction, pagination pkg.CursorPagination) FosterChildTransactionListResponse {
	var responses []FosterChildTransactionResponse
	for _, t := range transactions {
		responses = append(responses, t.toFosterChildTransactionResponse())
	}
	if responses == nil {
		responses = []FosterChildTransactionResponse{}
	}
	return FosterChildTransactionListResponse{
		Transactions: responses,
		Pagination:   pagination,
	}
}
