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
	pkg.PaginationParams
}

type DonationProgramExpenseExportParams struct {
	StartDate string `form:"start_date"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"end_date"`   // optional, format: YYYY-MM-DD
}
