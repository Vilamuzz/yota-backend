package social_program_expense

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramExpenseRequest struct {
	Title       string                `form:"title"`
	Amount      float64               `form:"amount"`
	ExpenseDate time.Time             `form:"expenseDate" time_format:"2006-01-02"`
	Note        string                `form:"note"`
	ProofFile   *multipart.FileHeader `form:"proofFile" swaggerignore:"true"`
}

type SocialProgramExpenseQueryParams struct {
	Search    string `form:"search"`
	SortBy    string `form:"sortBy"`
	StartDate string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"endDate"`   // optional, format: YYYY-MM-DD
	pkg.PaginationParams
}

type SocialProgramExpenseExportParams struct {
	StartDate string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"endDate"`   // optional, format: YYYY-MM-DD
}
