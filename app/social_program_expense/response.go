package social_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	ProofFile   string    `json:"proofFile"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SocialProgramExpenseDetailResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SocialProgramExpenseListResponse struct {
	SocialProgramExpenses []SocialProgramExpenseResponse `json:"expenses"`
	Pagination            pkg.CursorPagination           `json:"pagination"`
}

func (r *SocialProgramExpense) toSocialProgramExpenseDetailResponse() SocialProgramExpenseDetailResponse {
	return SocialProgramExpenseDetailResponse{
		ID:          r.ID.String(),
		Title:       r.Title,
		Amount:      r.Amount,
		ExpenseDate: r.ExpenseDate,
		Note:        r.Note,
		CreatedAt:   r.CreatedAt,
	}
}

func (r *SocialProgramExpense) toSocialProgramExpenseResponse() SocialProgramExpenseResponse {
	return SocialProgramExpenseResponse{
		ID:          r.ID.String(),
		Title:       r.Title,
		Amount:      r.Amount,
		ExpenseDate: r.ExpenseDate,
		ProofFile:   r.ProofFile,
		CreatedAt:   r.CreatedAt,
	}
}

func toSocialProgramExpenseListResponse(expenses []SocialProgramExpense, pagination pkg.CursorPagination) SocialProgramExpenseListResponse {
	var responses []SocialProgramExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.toSocialProgramExpenseResponse())
	}
	if responses == nil {
		responses = []SocialProgramExpenseResponse{}
	}
	return SocialProgramExpenseListResponse{
		SocialProgramExpenses: responses,
		Pagination:            pagination,
	}
}
