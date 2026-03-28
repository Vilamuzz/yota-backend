package foster_child_expense

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateFosterChildExpenseRequest struct {
	FosterChildID int     `json:"foster_child_id"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	Note          string  `json:"note"`
	ProofFile     string  `json:"proof_file"`
}

type UpdateFosterChildExpenseRequest struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Note        string  `json:"note"`
	ProofFile   string  `json:"proof_file"`
}

type FosterChildExpenseQueryParams struct {
	FosterChildID int    `form:"foster_child_id"`
	StartDate     string `form:"start_date"`
	EndDate       string `form:"end_date"`
	pkg.PaginationParams
}
