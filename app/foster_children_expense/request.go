package foster_children_expense

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type FosterChildrenExpenseRequest struct {
	Title       string                `form:"title"`
	Amount      float64               `form:"amount"`
	ExpenseDate time.Time             `form:"expenseDate" time_format:"2006-01-02"`
	Note        string                `form:"note"`
	ProofFile   *multipart.FileHeader `form:"proofFile" swaggerignore:"true"`
}

type FosterChildrenExpenseQueryParams struct {
	FosterChildrenID string `form:"fosterChildrenId"`
	Search           string `form:"search"`
	SortBy           string `form:"sortBy"`
	StartDate        string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate          string `form:"endDate"`   // optional, format: YYYY-MM-DD
	pkg.PaginationParams
}

type FosterChildrenExpenseExportParams struct {
	StartDate string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"endDate"`   // optional, format: YYYY-MM-DD
}
