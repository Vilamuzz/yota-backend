package foster_children_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateFosterChildrenTransactionRequest struct {
	DonorName   string  `json:"donorName"`
	DonorEmail  string  `json:"donorEmail"`
	GrossAmount float64 `json:"grossAmount"`
}

type FosterChildrenTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
