package donation_expense

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateExpenseRequest struct {
	DonationID string                `form:"donation_id" binding:"required"`
	Title      string                `form:"title" binding:"required"`
	Amount     float64               `form:"amount" binding:"required,gt=0"`
	Date       time.Time             `form:"date" binding:"required" time_format:"2006-01-02"`
	Note       string                `form:"note"`
	ProofFile  *multipart.FileHeader `form:"proof_file"`
}

type UpdateExpenseRequest struct {
	ID        string                `form:"id"`
	Title     string                `form:"title"`
	Amount    float64               `form:"amount"`
	Date      time.Time             `form:"date" time_format:"2006-01-02"`
	Note      string                `form:"note"`
	ProofFile *multipart.FileHeader `form:"proof_file"`
}

type DonationExpenseQueryParams struct {
	DonationID string `form:"donation_id"`
	pkg.PaginationParams
}
