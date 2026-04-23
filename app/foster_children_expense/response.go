package foster_children_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expense_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FosterChildrenExpenseDetailResponse struct {
	ID               string    `json:"id"`
	FosterChildrenID string    `json:"foster_children_id"`
	Title            string    `json:"title"`
	Amount           float64   `json:"amount"`
	ExpenseDate      time.Time `json:"expense_date"`
	Note             string    `json:"note"`
	ProofFile        string    `json:"proof_file"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FosterChildrenExpenseListResponse struct {
	Expenses   []FosterChildrenExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination            `json:"pagination"`
}

func (e *FosterChildrenExpense) toFosterChildrenExpenseDetailResponse() FosterChildrenExpenseDetailResponse {
	return FosterChildrenExpenseDetailResponse{
		ID:               e.ID.String(),
		FosterChildrenID: e.FosterChildrenID.String(),
		Title:            e.Title,
		Amount:           e.Amount,
		ExpenseDate:      e.ExpenseDate,
		Note:             e.Note,
		ProofFile:        e.ProofFile,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

func (e *FosterChildrenExpense) toFosterChildrenExpenseResponse() FosterChildrenExpenseResponse {
	return FosterChildrenExpenseResponse{
		ID:          e.ID.String(),
		Title:       e.Title,
		Amount:      e.Amount,
		ExpenseDate: e.ExpenseDate,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
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
