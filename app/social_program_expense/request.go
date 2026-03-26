package social_program_expense

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateSocialProgramExpenseRequest struct {
	DonationID string               `form:"donation_id"`
	Title      string               `form:"title"`
	Amount     float64              `form:"amount"`
	Date       time.Time            `form:"date" time_format:"2006-01-02"`
	Note       string               `form:"note"`
	ProofFile  multipart.FileHeader `form:"proof_file"`
}

type UpdateSocialProgramExpenseRequest struct {
	ID        string               `form:"id"`
	Title     string               `form:"title"`
	Amount    float64              `form:"amount"`
	Date      time.Time            `form:"date" time_format:"2006-01-02"`
	Note      string               `form:"note"`
	ProofFile multipart.FileHeader `form:"proof_file"`
}

type SocialProgramExpenseQueryParams struct {
	pkg.PaginationParams
}
