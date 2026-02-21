package transaction_donation

import "time"

type TransactionDonationResponse struct {
	ID              string     `json:"id"`
	DonationID      string     `json:"donation_id"`
	OrderID         string     `json:"order_id"`
	DonorName       string     `json:"donor_name"`
	DonorEmail      string     `json:"donor_email"`
	GrossAmount     float64    `json:"gross_amount"`
	PaymentMethod   string     `json:"payment_method"`
	PaymentStatus   string     `json:"payment_status"`
	Provider        string     `json:"provider"`
	TransactionID   string     `json:"transaction_id"`
	SnapToken       string     `json:"snap_token"`
	SnapRedirectURL string     `json:"snap_redirect_url"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

func toResponse(tx *TransactionDonation) TransactionDonationResponse {
	return TransactionDonationResponse{
		ID:              tx.ID,
		DonationID:      tx.DonationID,
		OrderID:         tx.OrderID,
		DonorName:       tx.DonorName,
		DonorEmail:      tx.DonorEmail,
		GrossAmount:     tx.GrossAmount,
		PaymentMethod:   tx.PaymentMethod,
		PaymentStatus:   tx.PaymentStatus,
		Provider:        tx.Provider,
		TransactionID:   tx.TransactionID,
		SnapToken:       tx.SnapToken,
		SnapRedirectURL: tx.SnapRedirectURL,
		PaidAt:          tx.PaidAt,
		CreatedAt:       tx.CreatedAt,
	}
}
