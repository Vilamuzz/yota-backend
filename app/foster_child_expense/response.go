package foster_child_expense

import "github.com/Vilamuzz/yota-backend/pkg"

type FosterChildExpenseResponse struct {
	ID            int     `json:"id"`
	FosterChildID int     `json:"foster_child_id"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
	Note          string  `json:"note"`
	ProofFile     string  `json:"proof_file"`
	CreatedAt     string  `json:"created_at"`
}

type FosterChildExpenseListResponse struct {
	Expenses   []FosterChildExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination         `json:"pagination"`
}

func (e *FosterChildExpense) ToFosterChildExpenseResponse() FosterChildExpenseResponse {
	return FosterChildExpenseResponse{
		ID:            e.ID,
		FosterChildID: e.FosterChildID,
		Title:         e.Title,
		Amount:        e.Amount,
		Date:          e.Date.Format("2006-01-02"),
		Note:          e.Note,
		ProofFile:     e.ProofFile,
		CreatedAt:     e.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToFosterChildExpenseListResponse(expenses []FosterChildExpense, pagination pkg.CursorPagination) FosterChildExpenseListResponse {
	var responses []FosterChildExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.ToFosterChildExpenseResponse())
	}
	if responses == nil {
		responses = []FosterChildExpenseResponse{}
	}
	return FosterChildExpenseListResponse{
		Expenses:   responses,
		Pagination: pagination,
	}
}
