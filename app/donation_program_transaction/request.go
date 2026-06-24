package donation_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateDonationProgramTransactionRequest struct {
	DonorName     string  `json:"donorName"`
	DonorEmail    string  `json:"donorEmail"`
	GrossAmount   float64 `json:"grossAmount"`
	PrayerContent string  `json:"prayerContent"`
}

type DonationProgramTransactionQueryParams struct {
	Search    string `form:"search"`
	Status    string `form:"status"`
	SortBy    string `form:"sortBy"`    // sort by gross amount, created at
	StartDate string `form:"startDate"` // optional, format: YYYY-MM-DD
	EndDate   string `form:"endDate"`   // optional, format: YYYY-MM-DD
	pkg.PaginationParams
}

type MonthlyIncomeQueryParams struct {
	Year string `form:"year"`
}
