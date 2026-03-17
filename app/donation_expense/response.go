package donation_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationExpenseResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DonationExpenseDetailResponse struct {
	ID         string    `json:"id"`
	DonationID string    `json:"donation_id"`
	Title      string    `json:"title"`
	Amount     float64   `json:"amount"`
	Date       time.Time `json:"date"`
	Note       string    `json:"note"`
	ProofFile  string    `json:"proof_file"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DonationExpenseListResponse struct {
	Expenses   []DonationExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination      `json:"pagination"`
}

func (r *DonationExpense) toDonationExpenseDetailResponse() DonationExpenseDetailResponse {
	return DonationExpenseDetailResponse{
		ID:         r.ID,
		DonationID: r.DonationID,
		Title:      r.Title,
		Amount:     r.Amount,
		Date:       r.Date,
		Note:       r.Note,
		ProofFile:  r.ProofFile,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (r *DonationExpense) toDonationExpenseResponse() DonationExpenseResponse {
	return DonationExpenseResponse{
		ID:        r.ID,
		Title:     r.Title,
		Amount:    r.Amount,
		Date:      r.Date,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func toDonationExpenseListResponse(expenses []DonationExpense, pagination pkg.CursorPagination) DonationExpenseListResponse {
	var responses []DonationExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.toDonationExpenseResponse())
	}
	if responses == nil {
		responses = []DonationExpenseResponse{}
	}
	return DonationExpenseListResponse{
		Expenses:   responses,
		Pagination: pagination,
	}
}
