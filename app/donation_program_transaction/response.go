package donation_program_transaction

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramTransactionResponse struct {
	ID                   string     `json:"id"`
	DonationProgramTitle string     `json:"donationProgramTitle"`
	OrderID              string     `json:"orderId"`
	DonorName            string     `json:"donorName"`
	DonorEmail           string     `json:"donorEmail"`
	IsOnline             bool       `json:"isOnline"`
	GrossAmount          float64    `json:"grossAmount"`
	TransactionStatus    string     `json:"transactionStatus"`
	SnapToken            string     `json:"snapToken"`
	PaidAt               *time.Time `json:"paidAt"`
	CreatedAt            time.Time  `json:"createdAt"`
}

type DonationProgramTransactionListResponse struct {
	Transactions []DonationProgramTransactionResponse `json:"transactions"`
	Pagination   pkg.CursorPagination                 `json:"pagination"`
}

func (tx *DonationProgramTransaction) toDonationProgramTransactionResponse() DonationProgramTransactionResponse {
	return DonationProgramTransactionResponse{
		ID:                   tx.ID.String(),
		DonationProgramTitle: tx.DonationProgram.Title,
		OrderID:              tx.OrderID,
		DonorName:            tx.DonorName,
		DonorEmail:           tx.DonorEmail,
		IsOnline:             tx.IsOnline,
		GrossAmount:          tx.GrossAmount,
		TransactionStatus:    tx.TransactionStatus,
		SnapToken:            tx.SnapToken,
		PaidAt:               tx.PaidAt,
		CreatedAt:            tx.CreatedAt,
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

type TransactionMonthlyIncomeItem struct {
	Month  string  `json:"month"`
	Income float64 `json:"income"`
}

type TransactionMonthlyIncomeRecord struct {
	DonationProgramID string                         `json:"donationProgramId"`
	Items             []TransactionMonthlyIncomeItem `json:"items"`
}
