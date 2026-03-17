package donation_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationTransactionResponse struct {
	ID                string     `json:"id"`
	DonationID        string     `json:"donation_id"`
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

type DonationTransactionListResponse struct {
	Transactions []DonationTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination          `json:"pagination"`
}

func (tx *DonationTransaction) toDonationTransactionResponse() DonationTransactionResponse {
	return DonationTransactionResponse{
		ID:                tx.ID,
		DonationID:        tx.DonationID,
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

func toDonationTransactionListResponse(transactions []DonationTransaction, pagination pkg.CursorPagination) DonationTransactionListResponse {
	var responses []DonationTransactionResponse
	for _, t := range transactions {
		responses = append(responses, t.toDonationTransactionResponse())
	}
	if responses == nil {
		responses = []DonationTransactionResponse{}
	}
	return DonationTransactionListResponse{
		Transactions: responses,
		Pagination:   pagination,
	}
}
