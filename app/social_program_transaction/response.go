package social_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramTransactionResponse struct {
	ID                     string     `json:"id"`
	SocialProgramInvoiceID string     `json:"socialProgramInvoiceId"`
	OrderID                string     `json:"orderId"`
	AccountID              string     `json:"accountId"`
	IsOnline               bool       `json:"isOnline"`
	GrossAmount            float64    `json:"grossAmount"`
	Fee                    float64    `json:"fee"`
	NetAmount              float64    `json:"netAmount"`
	PpnPercentage          float64    `json:"ppnPercentage"`
	PpnAmount              float64    `json:"ppnAmount"`
	TransactionStatus      string     `json:"transactionStatus"`
	Provider               string     `json:"provider"`
	TransactionID          string     `json:"transactionId"`
	SnapToken              string     `json:"snapToken"`
	PaidAt                 *time.Time `json:"paidAt"`
	CreatedAt              time.Time  `json:"createdAt"`
}

type SocialProgramTransactionListResponse struct {
	SocialProgramTransactions []SocialProgramTransactionResponse `json:"socialProgramTransactions"`
	Pagination                pkg.CursorPagination               `json:"pagination"`
}

func (tx *SocialProgramTransaction) toSocialProgramTransactionResponse() SocialProgramTransactionResponse {
	return SocialProgramTransactionResponse{
		ID:                     tx.ID.String(),
		SocialProgramInvoiceID: tx.SocialProgramInvoiceID.String(),
		OrderID:                tx.OrderID,
		AccountID:              tx.AccountID.String(),
		IsOnline:               tx.IsOnline,
		GrossAmount:            tx.GrossAmount,
		Fee:                    tx.Fee,
		NetAmount:              tx.NetAmount,
		PpnPercentage:          tx.PpnPercentage,
		PpnAmount:              tx.PpnAmount,
		TransactionStatus:      tx.TransactionStatus,
		Provider:               tx.Provider,
		TransactionID:          tx.TransactionID,
		SnapToken:              tx.SnapToken,
		PaidAt:                 tx.PaidAt,
		CreatedAt:              tx.CreatedAt,
	}
}

func toSocialProgramTransactionListResponse(transactions []SocialProgramTransaction, pagination pkg.CursorPagination) SocialProgramTransactionListResponse {
	var responses []SocialProgramTransactionResponse
	for _, t := range transactions {
		responses = append(responses, t.toSocialProgramTransactionResponse())
	}
	if responses == nil {
		responses = []SocialProgramTransactionResponse{}
	}
	return SocialProgramTransactionListResponse{
		SocialProgramTransactions: responses,
		Pagination:                pagination,
	}
}
