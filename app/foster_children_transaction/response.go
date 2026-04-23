package foster_children_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenTransactionResponse struct {
	ID                string     `json:"id"`
	FosterChildrenID  string     `json:"foster_children_id"`
	OrderID           string     `json:"order_id"`
	DonorName         string     `json:"donor_name"`
	DonorEmail        string     `json:"donor_email"`
	IsOnline          bool       `json:"is_online"`
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

type FosterChildrenTransactionListResponse struct {
	Transactions []FosterChildrenTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination                `json:"pagination"`
}

func (tx *FosterChildrenTransaction) toFosterChildrenTransactionResponse() FosterChildrenTransactionResponse {
	return FosterChildrenTransactionResponse{
		ID:                tx.ID.String(),
		FosterChildrenID:  tx.FosterChildrenID.String(),
		OrderID:           tx.OrderID,
		DonorName:         tx.DonorName,
		DonorEmail:        tx.DonorEmail,
		IsOnline:          tx.IsOnline,
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

func toFosterChildrenTransactionListResponse(transactions []FosterChildrenTransaction, pagination pkg.CursorPagination) FosterChildrenTransactionListResponse {
	var responses []FosterChildrenTransactionResponse
	for _, t := range transactions {
		responses = append(responses, t.toFosterChildrenTransactionResponse())
	}
	if responses == nil {
		responses = []FosterChildrenTransactionResponse{}
	}
	return FosterChildrenTransactionListResponse{
		Transactions: responses,
		Pagination:   pagination,
	}
}
