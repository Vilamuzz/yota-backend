package donation_transaction

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	CreateOfflineTransaction(ctx context.Context, req CreateTransactionRequest, actorID string) pkg.Response
	CreateTransaction(ctx context.Context, req CreateTransactionRequest, userID string) pkg.Response
	HandleNotification(ctx context.Context, notification MidtransNotificationRequest) pkg.Response
	ListTransactions(ctx context.Context, queryParams DonationTransactionQueryParams) pkg.Response
	ListMyTransactions(ctx context.Context, queryParams DonationTransactionQueryParams, userID string) pkg.Response
	GetTransactionByID(ctx context.Context, id string) pkg.Response
	GetMyTransactionByID(ctx context.Context, id, userID string) pkg.Response
}

type service struct {
	repo           Repository
	userRepo       user.Repository
	donationRepo   donation.Repository
	prayerRepo     prayer.Repository
	financeRepo    finance_record.Repository
	midtransClient payment_pkg.Client
	logService     app_log.Service
	timeout        time.Duration
}

func NewService(repo Repository, userRepo user.Repository, donationRepo donation.Repository, prayerRepo prayer.Repository, financeRepo finance_record.Repository, midtransClient payment_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:           repo,
		userRepo:       userRepo,
		donationRepo:   donationRepo,
		prayerRepo:     prayerRepo,
		financeRepo:    financeRepo,
		midtransClient: midtransClient,
		logService:     logService,
		timeout:        timeout,
	}
}

func (s *service) CreateOfflineTransaction(ctx context.Context, req CreateTransactionRequest, actorID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	errValidation := make(map[string]string)

	if req.DonationID == "" {
		errValidation["donation_id"] = "Donation ID is required"
	} else {
		_, err := s.donationRepo.FindOne(ctx, map[string]interface{}{"id": req.DonationID, "status": donation.StatusActive})
		if err != nil {
			errValidation["donation_id"] = "Donation not found"
		}
	}

	if req.GrossAmount <= 0 {
		errValidation["gross_amount"] = "Gross amount must be greater than 0"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	donorName := "anonymous"
	if req.DonorName != "" {
		donorName = req.DonorName
	}
	donorEmail := "anonymous@example.com"
	if req.DonorEmail != "" {
		donorEmail = req.DonorEmail
	}

	now := time.Now()

	transaction := &DonationTransaction{
		ID:                uuid.New().String(),
		DonationID:        req.DonationID,
		OrderID:           fmt.Sprintf("OFF-%s", uuid.New().String()),
		DonorName:         donorName,
		DonorEmail:        donorEmail,
		Source:            false,
		GrossAmount:       req.GrossAmount,
		FraudStatus:       "accept",
		TransactionStatus: "settlement",
		Provider:          "offline",
		PaidAt:            &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repo.Create(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation_transaction.service",
			"donation_id": req.DonationID,
		}).WithError(err).Error("failed to save offline transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to save offline transaction", nil, nil)
	}

	// Auto-create finance record (income)
	if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeDonation,
		FundID:          transaction.DonationID,
		SourceType:      finance_record.SourceTypeTransaction,
		SourceID:        transaction.ID,
		Amount:          transaction.GrossAmount,
		TransactionDate: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "donation_transaction.service",
			"transaction_id": transaction.ID,
		}).WithError(err).Warn("failed to create finance record for offline transaction")
	}

	s.logService.CreateLog(ctx, &actorID, "CREATE", "donation_transaction", transaction.ID, nil, transaction.toDonationTransactionResponse())

	return pkg.NewResponse(http.StatusCreated, "Offline transaction created successfully", nil, transaction.toDonationTransactionResponse())
}

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	if req.DonationID == "" {
		errValidation["donation_id"] = "Donation ID is required"
	} else {
		_, err := s.donationRepo.FindOne(ctx, map[string]interface{}{"id": req.DonationID, "status": donation.StatusActive})
		if err != nil {
			errValidation["donation_id"] = "Donation not found"
		}
	}

	if userID != "" {
		_, err := s.userRepo.FindOne(ctx, map[string]interface{}{"id": userID})
		if err != nil {
			errValidation["user_id"] = "User not found"
		}
	}

	if req.GrossAmount <= 0 {
		errValidation["gross_amount"] = "Gross amount must be greater than 0"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	donorName := "anonymous"
	if req.DonorName != "" {
		donorName = req.DonorName
	}
	donorEmail := "anonymous@example.com"
	if req.DonorEmail != "" {
		donorEmail = req.DonorEmail
	}

	orderID := fmt.Sprintf("DON-%s", uuid.New().String())
	grossAmountInt := int64(req.GrossAmount)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: grossAmountInt,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: donorName,
			Email: donorEmail,
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
	transaction := &DonationTransaction{
		ID:                uuid.New().String(),
		DonationID:        req.DonationID,
		UserID:            userID, // attach here
		OrderID:           orderID,
		DonorName:         donorName,
		DonorEmail:        donorEmail,
		Source:            true, // online
		GrossAmount:       req.GrossAmount,
		FraudStatus:       "accept",
		TransactionStatus: "pending",
		Provider:          "midtrans",
		SnapToken:         snapResp.Token,
		SnapRedirectURL:   snapResp.RedirectURL,
		PrayerContent:     req.PrayerContent,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Create(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation_transaction.service",
			"donation_id": req.DonationID,
			"order_id":    orderID,
		}).WithError(err).Error("failed to save online transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to save transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Transaction created successfully", nil, transaction.toDonationTransactionResponse())
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

	transaction, err := s.repo.FindByOrderID(ctx, notification.OrderID)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
	}

	if notification.TransactionStatus == transaction.TransactionStatus {
		return pkg.NewResponse(http.StatusOK, "No status change", nil, nil)
	}

	updates := map[string]interface{}{
		"transaction_status": notification.TransactionStatus,
		"fraud_status":       notification.FraudStatus,
		"updated_at":         time.Now(),
	}
	if notification.TransactionID != "" {
		updates["transaction_id"] = notification.TransactionID
	}
	isSettled := notification.TransactionStatus == "settlement" ||
		(notification.TransactionStatus == "capture" && notification.FraudStatus != "challenge")
	if isSettled {
		updates["paid_at"] = time.Now()
	}

	if err := s.repo.UpdateStatus(ctx, notification.OrderID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "donation_transaction.service",
			"transaction_id": transaction.ID,
			"order_id":       notification.OrderID,
		}).WithError(err).Error("failed to update transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update transaction", nil, nil)
	}

	if isSettled && transaction.PrayerContent != "" {
		now := time.Now()
		newPrayer := &prayer.Prayer{
			ID:         uuid.New().String(),
			DonationID: transaction.DonationID,
			UserID:     transaction.UserID,
			Content:    transaction.PrayerContent,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.prayerRepo.Create(ctx, newPrayer); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "donation_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       notification.OrderID,
			}).WithError(err).Warn("failed to create prayer after settlement")
		}
	}

	if isSettled {
		now := time.Now()
		if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
			ID:              uuid.New().String(),
			FundType:        finance_record.FundTypeDonation,
			FundID:          transaction.DonationID,
			SourceType:      finance_record.SourceTypeTransaction,
			SourceID:        transaction.ID,
			Amount:          transaction.GrossAmount,
			TransactionDate: now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "donation_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       notification.OrderID,
			}).WithError(err).Warn("failed to create finance record after settlement")
		}
		logrus.WithFields(logrus.Fields{
			"component":      "donation_transaction.service",
			"transaction_id": transaction.ID,
			"order_id":       notification.OrderID,
			"donation_id":    transaction.DonationID,
			"amount":         transaction.GrossAmount,
		}).Info("transaction settled")
	}

	return pkg.NewResponse(http.StatusOK, "Notification handled", nil, nil)
}

func (s *service) ListTransactions(ctx context.Context, params DonationTransactionQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.UserID != "" {
		options["user_id"] = params.UserID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	transactions, err := s.repo.FindAll(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation_transaction.service",
		}).WithError(err).Error("failed to fetch transactions")

		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transactions", nil, nil)
	}

	hasMore := len(transactions) > params.Limit
	if hasMore {
		transactions = transactions[:params.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
		}
	}

	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	var nextCursor, prevCursor string
	if len(transactions) > 0 {
		first := transactions[0]
		last := transactions[len(transactions)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toDonationTransactionListResponse(transactions, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}))
}

func (s *service) ListMyTransactions(ctx context.Context, params DonationTransactionQueryParams, userID string) pkg.Response {
	params.UserID = userID
	return s.ListTransactions(ctx, params)
}

func (s *service) GetTransactionByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid transaction ID format"}, nil)
	}

	transaction, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "donation_transaction.service",
			"transaction_id": id,
		}).WithError(err).Error("failed to fetch transaction")

		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, transaction.toDonationTransactionResponse())
}

func (s *service) GetMyTransactionByID(ctx context.Context, id, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid transaction ID format"}, nil)
	}

	transaction, err := s.repo.FindByIDAndUser(ctx, id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "donation_transaction.service",
			"transaction_id": id,
			"user_id":        userID,
		}).WithError(err).Error("failed to fetch transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, transaction.toDonationTransactionResponse())
}
