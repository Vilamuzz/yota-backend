package donation_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expense_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DonationProgramExpenseDetailResponse struct {
	ID                string    `json:"id"`
	DonationProgramID string    `json:"donation_program_id"`
	Title             string    `json:"title"`
	Amount            float64   `json:"amount"`
	ExpenseDate       time.Time `json:"expense_date"`
	Note              string    `json:"note"`
	ProofFile         string    `json:"proof_file"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type DonationProgramExpenseListResponse struct {
	Expenses   []DonationProgramExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination             `json:"pagination"`
}

func (r *DonationProgramExpense) toDonationProgramExpenseDetailResponse() DonationProgramExpenseDetailResponse {
	return DonationProgramExpenseDetailResponse{
		ID:                r.ID.String(),
		DonationProgramID: r.DonationProgramID.String(),
		Title:             r.Title,
		Amount:            r.Amount,
		ExpenseDate:       r.ExpenseDate,
		Note:              r.Note,
		ProofFile:         r.ProofFile,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func (r *DonationProgramExpense) toDonationProgramExpenseResponse() DonationProgramExpenseResponse {
	return DonationProgramExpenseResponse{
		ID:          r.ID.String(),
		Title:       r.Title,
		Amount:      r.Amount,
		ExpenseDate: r.ExpenseDate,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
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
