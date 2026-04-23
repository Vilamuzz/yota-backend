package social_program_invoice

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramInvoiceRequest struct {
	SubscriptionID string    `json:"subscription_id"`
	BillingPeriod  time.Time `json:"billing_period"`
	Amount         float64   `json:"amount"`
	Status         Status    `json:"status"`
	DueDate        time.Time `json:"due_date"`
}

type SocialProgramInvoiceQueryParams struct {
	SubscriptionID string `form:"subscription_id"`
	Status         string `form:"status"`
	pkg.PaginationParams
}
