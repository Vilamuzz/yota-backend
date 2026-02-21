package transaction_donation

type CreateTransactionRequest struct {
	DonationID  string  `json:"donation_id" binding:"required"`
	DonorName   string  `json:"donor_name" binding:"required"`
	DonorEmail  string  `json:"donor_email" binding:"required,email"`
	GrossAmount float64 `json:"gross_amount" binding:"required,gt=0"`
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

type QueryParams struct {
	Status     string `form:"status"`
	DonationID string `form:"donation_id"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
