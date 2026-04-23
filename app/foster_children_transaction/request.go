package foster_children_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateFosterChildrenTransactionRequest struct {
	DonorName   string  `json:"donor_name"`
	DonorEmail  string  `json:"donor_email"`
	GrossAmount float64 `json:"gross_amount"`
}


type FosterChildrenTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
