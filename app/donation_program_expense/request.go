package donation_program_expense

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramExpenseRequest struct {
	Title       string                `json:"title" form:"title"`
	Amount      float64               `json:"amount" form:"amount"`
	ExpenseDate string                `json:"expenseDate" form:"expenseDate"`
	Note        string                `json:"note" form:"note"`
	ProofFile   *multipart.FileHeader `form:"proofFile" swaggerignore:"true"`
}

type DonationProgramExpenseQueryParams struct {
	Search    string `form:"search"`
	SortBy    string `form:"sortBy"`
	StartDate string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"endDate"`   // optional, format: YYYY-MM-DD
	pkg.PaginationParams
}

type MonthlyExpenseQueryParams struct {
	Year string `form:"year"`
}
