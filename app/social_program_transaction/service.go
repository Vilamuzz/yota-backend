package social_program_transaction

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetSocialProgramTransactionList(ctx context.Context, accountID string, params SocialProgramTransactionQueryParams) pkg.Response
	GetSocialProgramTransactionByID(ctx context.Context, id string) pkg.Response
	CreateSocialProgramTransaction(ctx context.Context, accountID string, invoiceID string, payload CreateTransactionRequest) pkg.Response
	HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response
	GetMySocialProgramTransactionList(ctx context.Context, accountID string, params SocialProgramTransactionQueryParams) pkg.Response
	GetMySocialProgramTransactionByID(ctx context.Context, id string, accountID string) pkg.Response
	CreateOfflineSocialProgramTransaction(ctx context.Context, invoiceID string, payload CreateOfflineTransactionRequest) pkg.Response
}

type service struct {
	repo             Repository
	accountRepo      account.Repository
	subscriptionRepo social_program_subscription.Repository
	invoiceRepo      social_program_invoice.Repository
	financeRepo      finance_record.Repository
	midtransClient   payment_pkg.Client
	logService       app_log.Service
	timeout          time.Duration
}

func NewService(repo Repository, accountRepo account.Repository, subscriptionRepo social_program_subscription.Repository, invoiceRepo social_program_invoice.Repository, financeRepo finance_record.Repository, midtransClient payment_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:             repo,
		accountRepo:      accountRepo,
		subscriptionRepo: subscriptionRepo,
		invoiceRepo:      invoiceRepo,
		financeRepo:      financeRepo,
		midtransClient:   midtransClient,
		logService:       logService,
		timeout:          timeout,
	}
}

func (s *service) GetSocialProgramTransactionList(ctx context.Context, accountID string, params SocialProgramTransactionQueryParams) pkg.Response {
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
	if accountID != "" {
		options["account_id"] = accountID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	transactions, err := s.repo.FindAllSocialProgramTransactions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_transaction.service",
		}).WithError(err).Error("failed to fetch transactions")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data transaksi", nil, nil)
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

	return pkg.NewResponse(http.StatusOK, "Data transaksi berhasil ditemukan", nil, toSocialProgramTransactionListResponse(transactions, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetSocialProgramTransactionByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID transaksi tidak valid"}, nil)
	}

	transaction, err := s.repo.FindOneSocialProgramTransaction(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "social_program_transaction.service",
			"transaction_id": id,
		}).WithError(err).Error("failed to fetch transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data transaksi", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Data transaksi berhasil ditemukan", nil, transaction.toSocialProgramTransactionResponse())
}

func (s *service) CreateSocialProgramTransaction(ctx context.Context, accountID string, invoiceID string, payload CreateTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	invoice, err := s.invoiceRepo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{"id": invoiceID})
	if err != nil {
		errValidation["social_program_invoice_id"] = "Tagihan tidak ditemukan"
	} else if invoice.Status == social_program_invoice.InvoiceStatusPaid {
		return pkg.NewResponse(http.StatusBadRequest, "Tagihan sudah dibayar", nil, nil)
	}

	existingTx, err := s.repo.FindOneSocialProgramTransaction(ctx, map[string]interface{}{"social_program_invoice_id": invoiceID})
	if err == nil {
		if existingTx.TransactionStatus == "pending" {
			return pkg.NewResponse(http.StatusOK, "Menunggu pembayaran", nil, existingTx.toSocialProgramTransactionResponse())
		}
		if existingTx.TransactionStatus == "settlement" || existingTx.TransactionStatus == "capture" {
			return pkg.NewResponse(http.StatusBadRequest, "Tagihan sudah dibayar", nil, nil)
		}
	}

	account, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		errValidation["account_id"] = "Akun tidak ditemukan"
	}

	if payload.GrossAmount <= 0 {
		errValidation["gross_amount"] = "Jumlah nominal harus lebih besar dari 0"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	donorName := "anonymous"
	donorEmail := "anonymous@example.com"
	if account != nil {
		if account.UserProfile.Username != "" {
			donorName = account.UserProfile.Username
		}
		donorEmail = account.Email
	}

	orderID := fmt.Sprintf("SPI-%s", uuid.New().String())
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
				ID:    invoiceID,
				Price: grossAmountInt,
				Qty:   1,
				Name:  "Social Program Invoice Payment",
			},
		},
	}

	snapResp, err := s.midtransClient.CreateSnapTransaction(snapReq)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat transaksi Midtrans: "+err.Error(), nil, nil)
	}

	now := time.Now()
	transaction := &SocialProgramTransaction{
		ID:                     uuid.New(),
		SocialProgramInvoiceID: uuid.MustParse(invoiceID),
		AccountID:              uuid.MustParse(accountID),
		OrderID:                orderID,
		IsOnline:               true,
		GrossAmount:            payload.GrossAmount,
		FraudStatus:            "accept",
		TransactionStatus:      "pending",
		Provider:               "midtrans",
		SnapToken:              snapResp.Token,
		SnapRedirectURL:        snapResp.RedirectURL,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := s.repo.CreateSocialProgramTransaction(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_transaction.service",
			"invoice_id": invoiceID,
			"order_id":   orderID,
		}).WithError(err).Error("failed to save transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan transaksi", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Transaksi berhasil dibuat", nil, transaction.toSocialProgramTransactionResponse())
}

// HandleNotification processes payment notifications from Midtrans and updates transaction status accordingly
func (s *service) HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + s.midtransClient.GetServerKey()
	hash := sha512.Sum512([]byte(raw))
	expectedSig := fmt.Sprintf("%x", hash)
	if expectedSig != payload.SignatureKey {
		return pkg.NewResponse(http.StatusUnauthorized, "Tanda tangan tidak valid", nil, nil)
	}

	transaction, err := s.repo.FindOneSocialProgramTransaction(ctx, map[string]interface{}{"order_id": payload.OrderID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
	}

	if payload.TransactionStatus == transaction.TransactionStatus {
		return pkg.NewResponse(http.StatusOK, "Tidak ada perubahan status", nil, nil)
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
		now := time.Now()
		updates["paid_at"] = now

		_ = s.invoiceRepo.UpdateSocialProgramInvoice(ctx, transaction.SocialProgramInvoiceID.String(), map[string]interface{}{
			"status":     "paid",
			"updated_at": now,
		})

		invoice, err := s.invoiceRepo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{
			"id": transaction.SocialProgramInvoiceID.String(),
		})
		if err == nil {
			if err := s.subscriptionRepo.UpdateSocialProgramSubscription(ctx, invoice.SubscriptionID.String(), map[string]interface{}{
				"total_paid_periods": gorm.Expr("total_paid_periods + 1"),
				"updated_at":         now,
			}); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":       "social_program_transaction.service",
					"subscription_id": invoice.SubscriptionID,
				}).WithError(err).Error("failed to update subscription paid periods")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"component":  "social_program_transaction.service",
				"invoice_id": transaction.SocialProgramInvoiceID,
			}).WithError(err).Error("failed to find invoice for subscription update")
		}
	}

	if err := s.repo.UpdateSocialProgramTransaction(ctx, payload.OrderID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "social_program_transaction.service",
			"transaction_id": transaction.ID,
			"order_id":       payload.OrderID,
		}).WithError(err).Error("failed to update transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui transaksi", nil, nil)
	}

	if isSettled {
		now := time.Now()
		if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
			ID:              uuid.New().String(),
			FundType:        finance_record.FundTypeSocialProgram,
			FundID:          transaction.SocialProgramInvoiceID.String(),
			SourceType:      finance_record.SourceTypeTransaction,
			SourceID:        transaction.ID.String(),
			Amount:          transaction.GrossAmount,
			TransactionDate: now,
			CreatedAt:       now,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "social_program_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       payload.OrderID,
			}).WithError(err).Warn("failed to create finance record after settlement")
		}
	}

	return pkg.NewResponse(http.StatusOK, "Notifikasi berhasil ditangani", nil, nil)
}

func (s *service) GetMySocialProgramTransactionList(ctx context.Context, accountID string, params SocialProgramTransactionQueryParams) pkg.Response {
	return s.GetSocialProgramTransactionList(ctx, accountID, params)
}

func (s *service) GetMySocialProgramTransactionByID(ctx context.Context, id string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID transaksi tidak valid"}, nil)
	}

	transaction, err := s.repo.FindOneSocialProgramTransaction(ctx, map[string]interface{}{"id": id, "account_id": accountID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data transaksi", nil, nil)
	}

	if transaction.AccountID.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Akses ditolak", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Data transaksi berhasil ditemukan", nil, transaction.toSocialProgramTransactionResponse())
}

func (s *service) CreateOfflineSocialProgramTransaction(ctx context.Context, invoiceID string, payload CreateOfflineTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	invoice, err := s.invoiceRepo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{"id": invoiceID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Tagihan tidak ditemukan", nil, nil)
	}

	if invoice.Status == social_program_invoice.InvoiceStatusPaid {
		return pkg.NewResponse(http.StatusBadRequest, "Tagihan sudah dibayar", nil, nil)
	}

	now := time.Now()
	orderID := fmt.Sprintf("SPI-OFF-%s", uuid.New().String())

	transaction := &SocialProgramTransaction{
		ID:                     uuid.New(),
		SocialProgramInvoiceID: invoice.ID,
		AccountID:              invoice.Subscription.AccountID, // Assuming preloaded or we can get it from invoice.Subscription
		OrderID:                orderID,
		IsOnline:               false,
		GrossAmount:            payload.GrossAmount,
		FraudStatus:            "accept",
		TransactionStatus:      "settlement",
		Provider:               "offline",
		PaidAt:                 &now,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// Start transaction for atomicity
	err = s.repo.WithTransaction(ctx, func(txRepo Repository) error {
		if err := txRepo.CreateSocialProgramTransaction(ctx, transaction); err != nil {
			return err
		}

		if err := s.invoiceRepo.UpdateSocialProgramInvoice(ctx, invoiceID, map[string]interface{}{
			"status":     social_program_invoice.InvoiceStatusPaid,
			"updated_at": now,
		}); err != nil {
			return err
		}

		if err := s.subscriptionRepo.UpdateSocialProgramSubscription(ctx, invoice.SubscriptionID.String(), map[string]interface{}{
			"total_paid_periods": gorm.Expr("total_paid_periods + 1"),
			"updated_at":         now,
		}); err != nil {
			return err
		}

		if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
			ID:              uuid.New().String(),
			FundType:        finance_record.FundTypeSocialProgram,
			FundID:          invoice.ID.String(),
			SourceType:      finance_record.SourceTypeTransaction,
			SourceID:        transaction.ID.String(),
			Amount:          transaction.GrossAmount,
			TransactionDate: now,
			CreatedAt:       now,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_transaction.service",
			"invoice_id": invoiceID,
		}).WithError(err).Error("failed to create offline transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memproses pembayaran offline", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Pembayaran offline berhasil dicatat", nil, transaction.toSocialProgramTransactionResponse())
}
