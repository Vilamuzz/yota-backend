package foster_children_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenTransactionResponse struct {
	ID                 string     `json:"id"`
	FosterChildrenName string     `json:"fosterChildrenName"`
	OrderID            string     `json:"orderId"`
	DonorName          string     `json:"donorName"`
	DonorEmail         string     `json:"donorEmail"`
	IsOnline           bool       `json:"isOnline"`
	GrossAmount        float64    `json:"grossAmount"`
	TransactionStatus  string     `json:"transactionStatus"`
	TransactionID      string     `json:"transactionId"`
	SnapToken          string     `json:"snapToken"`
	PaidAt             *time.Time `json:"paidAt"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type FosterChildrenTransactionListResponse struct {
	Transactions []FosterChildrenTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination                `json:"pagination"`
}

func (tx *FosterChildrenTransaction) toFosterChildrenTransactionResponse() FosterChildrenTransactionResponse {
	return FosterChildrenTransactionResponse{
		ID:                 tx.ID.String(),
		FosterChildrenName: tx.FosterChildren.Name,
		OrderID:            tx.OrderID,
		DonorName:          tx.DonorName,
		DonorEmail:         tx.DonorEmail,
		IsOnline:           tx.IsOnline,
		GrossAmount:        tx.GrossAmount,
		TransactionStatus:  tx.TransactionStatus,
		TransactionID:      tx.TransactionID,
		SnapToken:          tx.SnapToken,
		PaidAt:             tx.PaidAt,
		CreatedAt:          tx.CreatedAt,
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
