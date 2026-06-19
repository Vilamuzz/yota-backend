package donation_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type DonationProgramExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	ProofFile   string    `json:"proofFile"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DonationProgramExpenseDetailResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DonationProgramExpenseListResponse struct {
	Expenses   []DonationProgramExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination             `json:"pagination"`
}

func (r *DonationProgramExpense) toDonationProgramExpenseResponse() DonationProgramExpenseResponse {
	return DonationProgramExpenseResponse{
		ID:          r.ID.String(),
		Title:       r.Title,
		Amount:      r.Amount,
		ExpenseDate: r.ExpenseDate,
		ProofFile:   s3_pkg.GetCDNURL(r.ProofFile),
		CreatedAt:   r.CreatedAt,
	}
}

func (r *DonationProgramExpense) toDonationProgramExpenseDetailResponse() DonationProgramExpenseDetailResponse {
	return DonationProgramExpenseDetailResponse{
		ID:          r.ID.String(),
		Title:       r.Title,
		Amount:      r.Amount,
		ExpenseDate: r.ExpenseDate,
		Note:        r.Note,
		CreatedAt:   r.CreatedAt,
	}
}

func toDonationProgramExpenseListResponse(expenses []DonationProgramExpense, pagination pkg.CursorPagination) DonationProgramExpenseListResponse {
	var responses []DonationProgramExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.toDonationProgramExpenseResponse())
	}
	if responses == nil {
		responses = []DonationProgramExpenseResponse{}
	}
	return DonationProgramExpenseListResponse{
		Expenses:   responses,
		Pagination: pagination,
	}
}

type MonthlyExpenseResponse struct {
	Month string  `json:"month"`
	Expense float64 `json:"expense"`
}

type MonthlyExpenseRecord struct {
	DonationProgramID string                 `json:"donationProgramId"`
	Items             []MonthlyExpenseResponse `json:"items"`
}