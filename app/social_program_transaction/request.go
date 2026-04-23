package social_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateTransactionRequest struct {
	SocialProgramInvoiceID string  `json:"social_program_invoice_id"`
	GrossAmount            float64 `json:"gross_amount"`
}


type SocialProgramTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
