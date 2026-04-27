package foster_children_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenExpenseResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	ExpenseDate time.Time `json:"expenseDate"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type FosterChildrenExpenseDetailResponse struct {
	ID               string    `json:"id"`
	FosterChildrenID string    `json:"fosterChildrenId"`
	Title            string    `json:"title"`
	Amount           float64   `json:"amount"`
	ExpenseDate      time.Time `json:"expenseDate"`
	Note             string    `json:"note"`
	ProofFile        string    `json:"proofFile"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type FosterChildrenExpenseListResponse struct {
	FosterChildrenExpenses []FosterChildrenExpenseResponse `json:"fosterChildrenExpenses"`
	Pagination             pkg.CursorPagination            `json:"pagination"`
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
		FosterChildrenExpenses: responses,
		Pagination:             pagination,
	}
}
