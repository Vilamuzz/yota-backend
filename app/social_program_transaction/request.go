package social_program_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateTransactionRequest struct {
	SocialProgramInvoiceID string  `json:"social_program_invoice_id"`
	UserID                 string  `json:"user_id"`
	GrossAmount            float64 `json:"gross_amount"`
}

type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
}

type SocialProgramTransactionQueryParams struct {
	Status string `form:"status"`
	pkg.PaginationParams
}
