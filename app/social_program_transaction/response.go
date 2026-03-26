package social_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramTransactionResponse struct {
	ID                     string     `json:"id"`
	SocialProgramInvoiceID string     `json:"social_program_invoice_id"`
	OrderID                string     `json:"order_id"`
	UserID                 string     `json:"user_id"`
	Source                 bool       `json:"source"`
	GrossAmount            float64    `json:"gross_amount"`
	FraudStatus            string     `json:"fraud_status"`
	TransactionStatus      string     `json:"transaction_status"`
	Provider               string     `json:"provider"`
	TransactionID          string     `json:"transaction_id"`
	SnapToken              string     `json:"snap_token"`
	SnapRedirectURL        string     `json:"snap_redirect_url"`
	PaidAt                 *time.Time `json:"paid_at"`
	CreatedAt              time.Time  `json:"created_at"`
}

type SocialProgramTransactionListResponse struct {
	Transactions []SocialProgramTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination               `json:"pagination"`
}

func (tx *SocialProgramTransaction) toSocialProgramTransactionResponse() SocialProgramTransactionResponse {
	return SocialProgramTransactionResponse{
		ID:                     tx.ID,
		SocialProgramInvoiceID: tx.SocialProgramInvoiceID,
		OrderID:                tx.OrderID,
		UserID:                 tx.UserID,
		Source:                 tx.Source,
		GrossAmount:            tx.GrossAmount,
		FraudStatus:            tx.FraudStatus,
		TransactionStatus:      tx.TransactionStatus,
		Provider:               tx.Provider,
		TransactionID:          tx.TransactionID,
		SnapToken:              tx.SnapToken,
		SnapRedirectURL:        tx.SnapRedirectURL,
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
		Transactions: responses,
		Pagination:   pagination,
	}
}
