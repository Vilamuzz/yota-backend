package donation_expense

import (
	"mime/multipart"
	"time"
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

type QueryParams struct {
	DonationID string `form:"donation_id"`
	NextCursor string `form:"next_cursor"`
	PrevCursor string `form:"prev_cursor"`
	Limit      int    `form:"limit"`
}
