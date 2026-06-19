package foster_children_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type FosterChildrenExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	ProofFile   string    `json:"proofFile"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FosterChildrenExpenseDetailResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FosterChildrenExpenseListResponse struct {
	Expenses   []FosterChildrenExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination            `json:"pagination"`
}

func (e *FosterChildrenExpense) toFosterChildrenExpenseDetailResponse() FosterChildrenExpenseDetailResponse {
	return FosterChildrenExpenseDetailResponse{
		ID:          e.ID.String(),
		Title:       e.Title,
		Amount:      e.Amount,
		ExpenseDate: e.ExpenseDate,
		Note:        e.Note,
		CreatedAt:   e.CreatedAt,
	}
}

func (e *FosterChildrenExpense) toFosterChildrenExpenseResponse() FosterChildrenExpenseResponse {
	return FosterChildrenExpenseResponse{
		ID:          e.ID.String(),
		Title:       e.Title,
		Amount:      e.Amount,
		ExpenseDate: e.ExpenseDate,
		ProofFile:   s3_pkg.GetCDNURL(e.ProofFile),
		CreatedAt:   e.CreatedAt,
	}
}

func toFosterChildrenExpenseListResponse(expenses []FosterChildrenExpense, pagination pkg.CursorPagination) FosterChildrenExpenseListResponse {
	var responses []FosterChildrenExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.toFosterChildrenExpenseResponse())
	}
	if responses == nil {
		responses = []FosterChildrenExpenseResponse{}
	}
	return FosterChildrenExpenseListResponse{
		Expenses:   responses,
		Pagination: pagination,
	}
}
