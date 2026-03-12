package donation_expense

import "time"

type ExpenseResponse struct {
	ID         string    `json:"id"`
	DonationID string    `json:"donation_id"`
	Title      string    `json:"title"`
	Amount     float64   `json:"amount"`
	Date       time.Time `json:"date"`
	Note       string    `json:"note"`
	ProofFile  string    `json:"proof_file"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
