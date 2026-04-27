package donation_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramTransactionResponse struct {
	ID                string     `json:"id"`
	DonationProgramID string     `json:"donationProgramId"`
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

type DonationProgramTransactionListResponse struct {
	DonationProgramTransactions []DonationProgramTransactionResponse `json:"donationProgramTransactions"`
	Pagination                  pkg.CursorPagination                 `json:"pagination"`
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
		DonationProgramTransactions: responses,
		Pagination:                  pagination,
	}
}
