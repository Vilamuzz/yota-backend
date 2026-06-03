package social_program_invoice

import (
	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramInvoiceQueryParams struct {
	SubscriptionID string `form:"subscriptionId"`
	Status         string `form:"status"`
	AccountID      string `json:"-"`
	pkg.PaginationParams
}
