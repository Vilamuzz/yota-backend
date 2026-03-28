package foster_child_expense

import "time"

type FosterChildExpense struct {
	ID            int       `json:"id"`
	FosterChildID int       `json:"foster_child_id"`
	Title         string    `json:"title"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
	Note          string    `json:"note"`
	ProofFile     string    `json:"proof_file"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
