package foster_child_transaction

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateFosterChildTransactionRequest struct {
	FosterChildID int     `json:"foster_child_id"`
	DonorName     string  `json:"donor_name"`
	DonorEmail    string  `json:"donor_email"`
	GrossAmount   float64 `json:"gross_amount"`
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

type FosterChildTransactionQueryParams struct {
	Status        string `form:"status"`
	FosterChildID string `form:"foster_child_id"`
	UserID        string `form:"user_id"`
	pkg.PaginationParams
}
