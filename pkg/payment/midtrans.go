package payment

import (
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

// Client defines the interface for Midtrans payment operations.
type Client interface {
	CreateSnapTransaction(req *snap.Request) (*snap.Response, error)
	GetServerKey() string
}

type midtransClient struct {
	snapClient snap.Client
	serverKey  string
}

// NewClient creates a new Midtrans Snap client from environment variables.
func NewClient() Client {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	env := midtrans.Sandbox
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		env = midtrans.Production
	}

	var c snap.Client
	c.New(serverKey, env)

	return &midtransClient{
		snapClient: c,
		serverKey:  serverKey,
	}
}

// CreateSnapTransaction creates a Midtrans Snap transaction and returns the token + redirect URL.
func (m *midtransClient) CreateSnapTransaction(req *snap.Request) (*snap.Response, error) {
	resp, err := m.snapClient.CreateTransaction(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetServerKey returns the server key used for signature verification.
func (m *midtransClient) GetServerKey() string {
	return m.serverKey
}

// MidtransNotificationRequest represents the notification payload from Midtrans.
type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
	VANumbers         []struct {
		Bank     string `json:"bank"`
		VANumber string `json:"va_number"`
	} `json:"va_numbers"`
	PermataVANumber string `json:"permata_va_number"`
}

// GetPaymentMethodCode resolves a Midtrans notification payload to an internal
// payment method code used to look up fee configuration.
func GetPaymentMethodCode(payload MidtransNotificationRequest) string {
	switch payload.PaymentType {
	case "gopay":
		return "gopay"
	case "qris":
		return "qris"
	case "shopeepay":
		return "shopeepay"
	case "echannel":
		return "mandiri_va"
	case "bank_transfer":
		if len(payload.VANumbers) > 0 {
			bank := payload.VANumbers[0].Bank
			switch bank {
			case "bca":
				return "bca_va"
			case "bri":
				return "bri_va"
			case "bni":
				return "bni_va"
			case "cimb":
				return "cimb_va"
			}
		}
		if payload.PermataVANumber != "" {
			return "permata_va"
		}
	case "permata":
		return "permata_va"
	case "dana":
		return "dana"
	case "ovo":
		return "ovo"
	}
	return ""
}
