package donation_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramTransactionResponse struct {
	ID                string     `json:"id"`
	DonationProgramID string     `json:"donation_program_id"`
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

type DonationProgramTransactionListResponse struct {
	Transactions []DonationProgramTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination                 `json:"pagination"`
}

func (tx *DonationProgramTransaction) toDonationProgramTransactionResponse() DonationProgramTransactionResponse {
	return DonationProgramTransactionResponse{
		ID:                tx.ID.String(),
		DonationProgramID: tx.DonationProgramID.String(),
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

func toDonationTransactionListResponse(transactions []DonationProgramTransaction, pagination pkg.CursorPagination) DonationProgramTransactionListResponse {
	var responses []DonationProgramTransactionResponse
	for _, t := range transactions {
		responses = append(responses, t.toDonationProgramTransactionResponse())
	}
	if responses == nil {
		responses = []DonationProgramTransactionResponse{}
	}
	return DonationProgramTransactionListResponse{
		Transactions: responses,
		Pagination:   pagination,
	}
}