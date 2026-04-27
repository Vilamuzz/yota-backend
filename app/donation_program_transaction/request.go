package donation_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateDonationProgramTransactionRequest struct {
	DonorName     string  `json:"donorName"`
	DonorEmail    string  `json:"donorEmail"`
	GrossAmount   float64 `json:"grossAmount"`
	PrayerContent string  `json:"prayerContent"`
}

type DonationProgramTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}

type PrayerQueryParams struct {
	Reported bool `form:"reported"`
	pkg.PaginationParams
}

type ReportPrayerRequest struct {
	Reason string `json:"reason"`
}
