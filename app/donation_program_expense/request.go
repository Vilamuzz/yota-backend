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
