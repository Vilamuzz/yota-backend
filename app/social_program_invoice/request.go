package social_program_invoice

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramInvoiceRequest struct {
	SubscriptionID string    `json:"subscriptionId"`
	BillingPeriod  time.Time `json:"billingPeriod"`
	Amount         float64   `json:"amount"`
	Status         Status    `json:"status"`
	DueDate        time.Time `json:"dueDate"`
}

type SocialProgramInvoiceQueryParams struct {
	SubscriptionID string `form:"subscriptionId"`
	Status         string `form:"status"`
	pkg.PaginationParams
}
