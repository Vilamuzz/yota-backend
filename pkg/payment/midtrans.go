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
