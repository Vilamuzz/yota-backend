package social_program_expense

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramExpenseResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SocialProgramExpenseDetailResponse struct {
	ID              string    `json:"id"`
	SocialProgramID string    `json:"social_program_id"`
	Title           string    `json:"title"`
	Amount          float64   `json:"amount"`
	Date            time.Time `json:"date"`
	Note            string    `json:"note"`
	ProofFile       string    `json:"proof_file"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SocialProgramExpenseListResponse struct {
	Expenses   []SocialProgramExpenseResponse `json:"expenses"`
	Pagination pkg.CursorPagination           `json:"pagination"`
}

func (r *SocialProgramExpense) toSocialProgramExpenseDetailResponse() SocialProgramExpenseDetailResponse {
	return SocialProgramExpenseDetailResponse{
		ID:              r.ID,
		SocialProgramID: r.SocialProgramID,
		Title:           r.Title,
		Amount:          r.Amount,
		Date:            r.Date,
		Note:            r.Note,
		ProofFile:       r.ProofFile,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func (r *SocialProgramExpense) toSocialProgramExpenseResponse() SocialProgramExpenseResponse {
	return SocialProgramExpenseResponse{
		ID:        r.ID,
		Title:     r.Title,
		Amount:    r.Amount,
		Date:      r.Date,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func toSocialProgramExpenseListResponse(expenses []SocialProgramExpense, pagination pkg.CursorPagination) SocialProgramExpenseListResponse {
	var responses []SocialProgramExpenseResponse
	for _, expense := range expenses {
		responses = append(responses, expense.toSocialProgramExpenseResponse())
	}
	if responses == nil {
		responses = []SocialProgramExpenseResponse{}
	}
	return SocialProgramExpenseListResponse{
		Expenses:   responses,
		Pagination: pagination,
	}
}
