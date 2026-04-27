package donation_program_expense

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramExpenseRequest struct {
	Title       string                `form:"title"`
	Amount      float64               `form:"amount"`
	ExpenseDate time.Time             `form:"expenseDate" time_format:"2006-01-02"`
	Note        string                `form:"note"`
	ProofFile   *multipart.FileHeader `form:"proofFile" swaggerignore:"true"`
}

type DonationProgramExpenseQueryParams struct {
	pkg.PaginationParams
}
