package transaction_donation

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) pkg.Response
	HandleNotification(ctx context.Context, notification MidtransNotificationRequest) pkg.Response
	List(ctx context.Context, params QueryParams) pkg.Response
	GetByID(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo           Repository
	midtransClient payment_pkg.Client
	timeout        time.Duration
}

func NewService(repo Repository, midtransClient payment_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:           repo,
		midtransClient: midtransClient,
		timeout:        timeout,
	}
}

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	orderID := fmt.Sprintf("DON-%s", uuid.New().String())
	grossAmountInt := int64(req.GrossAmount)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: grossAmountInt,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.DonorName,
			Email: req.DonorEmail,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    req.DonationID,
				Price: grossAmountInt,
				Qty:   1,
				Name:  "Donation",
			},
		},
	}

	snapResp, err := s.midtransClient.CreateSnapTransaction(snapReq)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create Midtrans transaction: "+err.Error(), nil, nil)
	}

	now := time.Now()
	tx := &TransactionDonation{
		ID:              uuid.New().String(),
		DonationID:      req.DonationID,
		OrderID:         orderID,
		DonorName:       req.DonorName,
		DonorEmail:      req.DonorEmail,
		Source:          true, // online
		GrossAmount:     req.GrossAmount,
		PaymentStatus:   "pending",
		Provider:        "midtrans",
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to save transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Transaction created successfully", nil, toResponse(tx))
}

func (s *service) HandleNotification(ctx context.Context, notification MidtransNotificationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Verify signature key: SHA512(order_id + status_code + gross_amount + server_key)
	raw := notification.OrderID + notification.StatusCode + notification.GrossAmount + s.midtransClient.GetServerKey()
	hash := sha512.Sum512([]byte(raw))
	expectedSig := fmt.Sprintf("%x", hash)
	if expectedSig != notification.SignatureKey {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid signature", nil, nil)
	}

	// Find the transaction
	tx, err := s.repo.FindByOrderID(ctx, notification.OrderID)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
	}

	// Determine new payment status
	status := mapMidtransStatus(notification.TransactionStatus, notification.FraudStatus)
	if status == tx.PaymentStatus {
		return pkg.NewResponse(http.StatusOK, "No status change", nil, nil)
	}

	var paidAt *time.Time
	if status == "settlement" {
		now := time.Now()
		paidAt = &now
	}

	if err := s.repo.UpdateStatus(ctx, notification.OrderID, status, notification.TransactionID, paidAt); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Notification handled", nil, nil)
}

func (s *service) List(ctx context.Context, params QueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	options := map[string]interface{}{
		"limit":       params.Limit,
		"status":      params.Status,
		"donation_id": params.DonationID,
	}

	transactions, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transactions", nil, nil)
	}

	hasNext := len(transactions) > params.Limit
	if hasNext {
		transactions = transactions[:params.Limit]
	}

	responses := make([]TransactionDonationResponse, len(transactions))
	for i, tx := range transactions {
		txCopy := tx
		responses[i] = toResponse(&txCopy)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"transactions": responses,
		"has_next":     hasNext,
		"limit":        params.Limit,
	})
}

func (s *service) GetByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid transaction ID format"}, nil)
	}

	tx, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toResponse(tx))
}

// mapMidtransStatus normalizes Midtrans transaction_status + fraud_status into a simple status string.
func mapMidtransStatus(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			return "challenge"
		}
		return "settlement"
	case "settlement":
		return "settlement"
	case "deny", "cancel", "expire":
		return transactionStatus
	case "pending":
		return "pending"
	default:
		return transactionStatus
	}
}
