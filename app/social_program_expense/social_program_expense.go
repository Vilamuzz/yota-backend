package social_program_expense

import "time"

type SocialProgramExpense struct {
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
