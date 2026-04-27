package social_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateTransactionRequest struct {
	SocialProgramInvoiceID string  `json:"socialProgramInvoiceId"`
	GrossAmount            float64 `json:"grossAmount"`
}


type SocialProgramTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
