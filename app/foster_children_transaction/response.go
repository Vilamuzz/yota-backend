package foster_children_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenTransactionResponse struct {
	ID                string     `json:"id"`
	FosterChildrenID  string     `json:"fosterChildrenId"`
	OrderID           string     `json:"orderId"`
	DonorName         string     `json:"donorName"`
	DonorEmail        string     `json:"donorEmail"`
	IsOnline          bool       `json:"isOnline"`
	GrossAmount       float64    `json:"grossAmount"`
	FraudStatus       string     `json:"fraudStatus"`
	TransactionStatus string     `json:"transactionStatus"`
	Provider          string     `json:"provider"`
	TransactionID     string     `json:"transactionId"`
	SnapToken         string     `json:"snapToken"`
	SnapRedirectURL   string     `json:"snapRedirectUrl"`
	PaidAt            *time.Time `json:"paidAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type FosterChildrenTransactionListResponse struct {
	FosterChildrenTransactions []FosterChildrenTransactionResponse `json:"fosterChildrenTransactions"`
	Pagination                 pkg.CursorPagination                `json:"pagination"`
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
		FosterChildrenTransactions: responses,
		Pagination:                 pagination,
	}
}
