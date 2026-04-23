package donation_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateDonationProgramTransactionRequest struct {
	DonorName     string  `json:"donor_name"`
	DonorEmail    string  `json:"donor_email"`
	GrossAmount   float64 `json:"gross_amount"`
	PrayerContent string  `json:"prayer_content"`
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
