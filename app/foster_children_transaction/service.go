package foster_children_transaction

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetFosterChildrenTransactionList(ctx context.Context, accountID, fosterChildrenID string, params FosterChildrenTransactionQueryParams) pkg.Response
	GetFosterChildrenTransactionByID(ctx context.Context, fosterChildrenTransactionID string) pkg.Response
	CreateOfflineFosterChildrenTransaction(ctx context.Context, accountID, fosterChildrenID string, payload CreateFosterChildrenTransactionRequest) pkg.Response
	CreateFosterChildrenTransaction(ctx context.Context, accountID, fosterChildrenID string, payload CreateFosterChildrenTransactionRequest) pkg.Response
	HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response
	GetMyFosterChildrenTransactionList(ctx context.Context, accountID string, params FosterChildrenTransactionQueryParams) pkg.Response
	GetMyFosterChildrenTransactionByID(ctx context.Context, fosterChildrenTransactionID, accountID string) pkg.Response
}

type service struct {
	repo               Repository
	accountRepo        account.Repository
	fosterChildrenRepo foster_children.Repository
	financeRepo        finance_record.Repository
	midtransClient     payment_pkg.Client
	logService         app_log.Service
	timeout            time.Duration
}

func NewService(repo Repository, accountRepo account.Repository, fosterChildrenRepo foster_children.Repository, financeRepo finance_record.Repository, midtransClient payment_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:               repo,
		accountRepo:        accountRepo,
		fosterChildrenRepo: fosterChildrenRepo,
		financeRepo:        financeRepo,
		midtransClient:     midtransClient,
		logService:         logService,
		timeout:            timeout,
	}
}

func (s *service) GetFosterChildrenTransactionList(ctx context.Context, accountID, fosterChildrenID string, params FosterChildrenTransactionQueryParams) pkg.Response {
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
	if fosterChildrenID != "" {
		options["foster_children_id"] = fosterChildrenID
	}
	if accountID != "" {
		options["account_id"] = accountID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	transactions, err := s.repo.FindAllFosterChildrenTransactions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "foster_children_transaction.service",
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
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toFosterChildrenTransactionListResponse(transactions, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetFosterChildrenTransactionByID(ctx context.Context, fosterChildrenTransactionID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(fosterChildrenTransactionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid transaction ID format"}, nil)
	}

	transaction, err := s.repo.FindOneFosterChildrenTransaction(ctx, map[string]interface{}{"id": fosterChildrenTransactionID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "foster_children_transaction.service",
			"transaction_id": fosterChildrenTransactionID,
		}).WithError(err).Error("failed to fetch transaction")

		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, transaction.toFosterChildrenTransactionResponse())
}

func (s *service) CreateOfflineFosterChildrenTransaction(ctx context.Context, accountID, fosterChildrenID string, payload CreateFosterChildrenTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	errValidation := make(map[string]string)

	if fosterChildrenID == "" {
		errValidation["foster_children_id"] = "Foster Children ID is required"
	} else {
		_, err := s.fosterChildrenRepo.FindOneFosterChildren(ctx, map[string]interface{}{"id": fosterChildrenID})
		if err != nil {
			errValidation["foster_children_id"] = "Foster Children not found"
		}
	}

	if payload.GrossAmount <= 0 {
		errValidation["gross_amount"] = "Gross amount must be greater than 0"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	donorName := "anonymous"
	if payload.DonorName != "" {
		donorName = payload.DonorName
	}
	donorEmail := "anonymous@example.com"
	if payload.DonorEmail != "" {
		donorEmail = payload.DonorEmail
	}

	now := time.Now()

	transaction := &FosterChildrenTransaction{
		ID:                uuid.New(),
		FosterChildrenID:  uuid.MustParse(fosterChildrenID),
		OrderID:           fmt.Sprintf("OFF-%s", uuid.New().String()),
		DonorName:         donorName,
		DonorEmail:        donorEmail,
		IsOnline:          false,
		GrossAmount:       payload.GrossAmount,
		FraudStatus:       "accept",
		TransactionStatus: "settlement",
		Provider:          "offline",
		PaidAt:            &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repo.CreateFosterChildrenTransaction(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":          "foster_children_transaction.service",
			"foster_children_id": fosterChildrenID,
		}).WithError(err).Error("failed to save offline transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to save offline transaction", nil, nil)
	}

	// Auto-create finance record (income)
	if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeFosterChildren,
		FundID:          transaction.FosterChildrenID.String(),
		SourceType:      finance_record.SourceTypeTransaction,
		SourceID:        transaction.ID.String(),
		Amount:          transaction.GrossAmount,
		TransactionDate: now,
		CreatedAt:       now,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "foster_children_transaction.service",
			"transaction_id": transaction.ID,
		}).WithError(err).Warn("failed to create finance record for offline transaction")
	}

	s.logService.CreateLog(ctx, &accountID, "CREATE", "foster_children_transaction", transaction.ID.String(), nil, transaction.toFosterChildrenTransactionResponse())

	return pkg.NewResponse(http.StatusCreated, "Offline transaction created successfully", nil, transaction.toFosterChildrenTransactionResponse())
}

func (s *service) CreateFosterChildrenTransaction(ctx context.Context, accountID string, fosterChildrenID string, payload CreateFosterChildrenTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	if fosterChildrenID == "" {
		errValidation["foster_children_id"] = "Foster Children ID is required"
	} else {
		_, err := s.fosterChildrenRepo.FindOneFosterChildren(ctx, map[string]interface{}{"id": fosterChildrenID})
		if err != nil {
			errValidation["foster_children_id"] = "Foster Children not found"
		}
	}

	if accountID != "" {
		_, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
		if err != nil {
			errValidation["account_id"] = "Account not found"
		}
	}

	if payload.GrossAmount <= 0 {
		errValidation["gross_amount"] = "Gross amount must be greater than 0"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	donorName := "anonymous"
	if payload.DonorName != "" {
		donorName = payload.DonorName
	}
	donorEmail := "anonymous@example.com"
	if payload.DonorEmail != "" {
		donorEmail = payload.DonorEmail
	}

	orderID := fmt.Sprintf("FC-%s", uuid.New().String())
	grossAmountInt := int64(payload.GrossAmount)

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
				ID:    fosterChildrenID,
				Price: grossAmountInt,
				Qty:   1,
				Name:  "Foster Children Donation",
			},
		},
	}

	snapResp, err := s.midtransClient.CreateSnapTransaction(snapReq)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create Midtrans transaction: "+err.Error(), nil, nil)
	}

	now := time.Now()
	transaction := &FosterChildrenTransaction{
		ID:                uuid.New(),
		FosterChildrenID:  uuid.MustParse(fosterChildrenID),
		AccountID:         uuid.MustParse(accountID),
		OrderID:           orderID,
		DonorName:         donorName,
		DonorEmail:        donorEmail,
		IsOnline:          true,
		GrossAmount:       payload.GrossAmount,
		FraudStatus:       "accept",
		TransactionStatus: "pending",
		Provider:          "midtrans",
		SnapToken:         snapResp.Token,
		SnapRedirectURL:   snapResp.RedirectURL,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.CreateFosterChildrenTransaction(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":          "foster_children_transaction.service",
			"foster_children_id": fosterChildrenID,
			"order_id":           orderID,
		}).WithError(err).Error("failed to save online transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to save transaction", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Transaction created successfully", nil, transaction.toFosterChildrenTransactionResponse())
}

func (s *service) HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Verify signature key: SHA512(order_id + status_code + gross_amount + server_key)
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + s.midtransClient.GetServerKey()
	hash := sha512.Sum512([]byte(raw))
	expectedSig := fmt.Sprintf("%x", hash)
	if expectedSig != payload.SignatureKey {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid signature", nil, nil)
	}

	transaction, err := s.repo.FindOneFosterChildrenTransaction(ctx, map[string]interface{}{"order_id": payload.OrderID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
	}

	if payload.TransactionStatus == transaction.TransactionStatus {
		return pkg.NewResponse(http.StatusOK, "No status change", nil, nil)
	}

	updates := map[string]interface{}{
		"transaction_status": payload.TransactionStatus,
		"fraud_status":       payload.FraudStatus,
		"updated_at":         time.Now(),
	}
	if payload.TransactionID != "" {
		updates["transaction_id"] = payload.TransactionID
	}
	isSettled := payload.TransactionStatus == "settlement" ||
		(payload.TransactionStatus == "capture" && payload.FraudStatus != "challenge")
	if isSettled {
		updates["paid_at"] = time.Now()
	}

	if err := s.repo.UpdateFosterChildrenTransaction(ctx, payload.OrderID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "foster_children_transaction.service",
			"transaction_id": transaction.ID,
			"order_id":       payload.OrderID,
		}).WithError(err).Error("failed to update transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update transaction", nil, nil)
	}

	if isSettled {
		now := time.Now()
		if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
			ID:              uuid.New().String(),
			FundType:        finance_record.FundTypeFosterChildren,
			FundID:          transaction.FosterChildrenID.String(),
			SourceType:      finance_record.SourceTypeTransaction,
			SourceID:        transaction.ID.String(),
			Amount:          transaction.GrossAmount,
			TransactionDate: now,
			CreatedAt:       now,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "foster_children_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       payload.OrderID,
			}).WithError(err).Warn("failed to create finance record after settlement")
		}
		logrus.WithFields(logrus.Fields{
			"component":          "foster_children_transaction.service",
			"transaction_id":     transaction.ID,
			"order_id":           payload.OrderID,
			"foster_children_id": transaction.FosterChildrenID,
			"amount":             transaction.GrossAmount,
		}).Info("transaction settled")
	}

	return pkg.NewResponse(http.StatusOK, "Notification handled", nil, nil)
}

func (s *service) GetMyFosterChildrenTransactionList(ctx context.Context, accountID string, params FosterChildrenTransactionQueryParams) pkg.Response {
	return s.GetFosterChildrenTransactionList(ctx, accountID, "", params)
}

func (s *service) GetMyFosterChildrenTransactionByID(ctx context.Context, fosterChildrenTransactionID string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(fosterChildrenTransactionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid transaction ID format"}, nil)
	}

	transaction, err := s.repo.FindOneFosterChildrenTransaction(ctx, map[string]interface{}{"id": fosterChildrenTransactionID, "account_id": accountID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaction not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "foster_children_transaction.service",
			"transaction_id": fosterChildrenTransactionID,
			"account_id":     accountID,
		}).WithError(err).Error("failed to fetch transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch transaction", nil, nil)
	}

	if transaction.AccountID.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Forbidden", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, transaction.toFosterChildrenTransactionResponse())
}
