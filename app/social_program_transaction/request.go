package social_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateTransactionRequest struct {
	GrossAmount float64 `json:"grossAmount"`
}

type CreateOfflineTransactionRequest struct {
	GrossAmount float64 `json:"grossAmount" binding:"required,gt=0"`
}

type SocialProgramTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
