package donation_program_transaction

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetDonationProgramTransactionList(ctx context.Context, accountID, donationProgramID string, params DonationProgramTransactionQueryParams) pkg.Response
	GetDonationProgramTransactionByID(ctx context.Context, donationProgramTransactionID string) pkg.Response
	CreateOfflineDonationProgramTransaction(ctx context.Context, accountID, donationProgramID string, payload CreateDonationProgramTransactionRequest) pkg.Response
	CreateDonationProgramTransaction(ctx context.Context, accountID, donationSlug string, payload CreateDonationProgramTransactionRequest) pkg.Response
	CancelOfflineDonationProgramTransaction(ctx context.Context, transactionID string) pkg.Response

	HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response

	GetMyDonationProgramTransactionList(ctx context.Context, accountID string, params DonationProgramTransactionQueryParams) pkg.Response
	GetMyDonationProgramTransactionByID(ctx context.Context, donationProgramTransactionID, accountID string) pkg.Response
}

type service struct {
	repo           Repository
	accountRepo    account.Repository
	donationRepo   donation_program.Repository
	prayerRepo     prayer.Repository
	financeRepo    finance_record.Repository
	midtransClient payment_pkg.Client
	logService     app_log.Service
	timeout        time.Duration
}

func NewService(repo Repository, accountRepo account.Repository, donationRepo donation_program.Repository, prayerRepo prayer.Repository, financeRepo finance_record.Repository, midtransClient payment_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:           repo,
		accountRepo:    accountRepo,
		donationRepo:   donationRepo,
		prayerRepo:     prayerRepo,
		financeRepo:    financeRepo,
		midtransClient: midtransClient,
		logService:     logService,
		timeout:        timeout,
	}
}

func (s *service) GetDonationProgramTransactionList(ctx context.Context, accountID, donationProgramID string, params DonationProgramTransactionQueryParams) pkg.Response {
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
	if donationProgramID != "" {
		options["donation_program_id"] = donationProgramID
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

	transactions, err := s.repo.FindAllDonationProgramTransactions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation_program_transaction.service",
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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toDonationTransactionListResponse(transactions, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetDonationProgramTransactionByID(ctx context.Context, donationProgramTransactionID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(donationProgramTransactionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID transaksi tidak valid"}, nil)
	}

	transaction, err := s.repo.FindOneDonationProgramTransaction(ctx, map[string]interface{}{"id": donationProgramTransactionID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "donation_program_transaction.service",
			"transaction_id": donationProgramTransactionID,
		}).WithError(err).Error("failed to fetch transaction")

		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data transaksi", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, transaction.toDonationProgramTransactionResponse())
}

func (s *service) CreateOfflineDonationProgramTransaction(ctx context.Context, accountID, donationProgramID string, payload CreateDonationProgramTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	errValidation := make(map[string]string)

	if donationProgramID == "" {
		errValidation["donationProgramId"] = "ID Program Donasi wajib diisi"
	} else {
		_, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"id": donationProgramID, "status": donation_program.StatusActive})
		if err != nil {
			errValidation["donationProgramId"] = "Program Donasi tidak ditemukan"
		}
	}

	if payload.GrossAmount <= 0 {
		errValidation["grossAmount"] = "Jumlah kotor harus lebih besar dari 0"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
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

	transaction := &DonationProgramTransaction{
		ID:                uuid.New(),
		DonationProgramID: uuid.MustParse(donationProgramID),
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
	if err := s.repo.CreateDonationProgramTransaction(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_transaction.service",
			"donation_program_id": donationProgramID,
		}).WithError(err).Error("failed to save offline transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan transaksi offline", nil, nil)
	}

	// Auto-create finance record (income)
	if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeDonation,
		FundID:          transaction.DonationProgramID.String(),
		SourceType:      finance_record.SourceTypeTransaction,
		SourceID:        transaction.ID.String(),
		Amount:          transaction.GrossAmount,
		TransactionDate: now,
		CreatedAt:       now,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "donation_program_transaction.service",
			"transaction_id": transaction.ID,
		}).WithError(err).Warn("failed to create finance record for offline transaction")
	}

	s.logService.CreateLog(ctx, &accountID, "CREATE", "donation_program_transaction", transaction.ID.String(), nil, transaction.toDonationProgramTransactionResponse())

	return pkg.NewResponse(http.StatusCreated, "Transaksi offline berhasil dibuat", nil, transaction.toDonationProgramTransactionResponse())
}

func (s *service) CreateDonationProgramTransaction(ctx context.Context, accountID string, donationSlug string, payload CreateDonationProgramTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	var donationProgramID string

	if donationSlug == "" {
		errValidation["donationProgramSlug"] = "Slug Program Donasi wajib diisi"
	} else {
		program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": donationSlug, "status": donation_program.StatusActive})
		if err != nil {
			errValidation["donationProgramSlug"] = "Program Donasi tidak ditemukan"
		} else {
			donationProgramID = program.ID.String()
		}
	}

	if accountID != "" {
		_, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
		if err != nil {
			errValidation["accountId"] = "Akun tidak ditemukan"
		}
	}

	if payload.GrossAmount <= 0 {
		errValidation["grossAmount"] = "Jumlah kotor harus lebih besar dari 0"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	donorName := "anonymous"
	if payload.DonorName != "" {
		donorName = payload.DonorName
	}
	donorEmail := "anonymous@example.com"
	if payload.DonorEmail != "" {
		donorEmail = payload.DonorEmail
	}

	orderID := fmt.Sprintf("DON-%s", uuid.New().String())
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
				ID:    donationProgramID,
				Price: grossAmountInt,
				Qty:   1,
				Name:  "Donation",
			},
		},
	}

	snapResp, err := s.midtransClient.CreateSnapTransaction(snapReq)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat transaksi Midtrans: "+err.Error(), nil, nil)
	}

	var accountIDPtr *uuid.UUID
	if accountID != "" {
		id := uuid.MustParse(accountID)
		accountIDPtr = &id
	}

	now := time.Now()
	transaction := &DonationProgramTransaction{
		ID:                uuid.New(),
		DonationProgramID: uuid.MustParse(donationProgramID),
		AccountID:         accountIDPtr,
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

	if err := s.repo.CreateDonationProgramTransaction(ctx, transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_transaction.service",
			"donation_program_id": donationProgramID,
			"order_id":            orderID,
		}).WithError(err).Error("failed to save online transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan transaksi", nil, nil)
	}

	if payload.PrayerContent != "" {
		newPrayer := &prayer.Prayer{
			ID:                           uuid.New(),
			DonationProgramTransactionID: transaction.ID,
			Content:                      payload.PrayerContent,
			IsPublished:                  false, // pending, will be published on settlement
			CreatedAt:                    now,
		}
		if err := s.prayerRepo.CreatePrayer(ctx, newPrayer); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "donation_program_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       transaction.OrderID,
			}).WithError(err).Warn("failed to create prayer")
		}
	}

	return pkg.NewResponse(http.StatusCreated, "Transaksi berhasil dibuat", nil, transaction.toDonationProgramTransactionResponse())
}

func (s *service) CancelOfflineDonationProgramTransaction(ctx context.Context, transactionID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	transaction, err := s.repo.FindOneDonationProgramTransaction(ctx, map[string]interface{}{"id": transactionID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
	}
	if transaction.IsOnline {
		return pkg.NewResponse(http.StatusBadRequest, "Transaksi online tidak dapat dibatalkan", nil, nil)
	}

	if err := s.repo.CancelDonationProgramTransaction(ctx, transactionID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membatalkan transaksi", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Transaksi berhasil dibatalkan", nil, nil)
}

func (s *service) HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Verify signature key: SHA512(order_id + status_code + gross_amount + server_key)
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + s.midtransClient.GetServerKey()
	hash := sha512.Sum512([]byte(raw))
	expectedSig := fmt.Sprintf("%x", hash)
	if expectedSig != payload.SignatureKey {
		return pkg.NewResponse(http.StatusUnauthorized, "Tanda tangan tidak valid", nil, nil)
	}

	transaction, err := s.repo.FindOneDonationProgramTransaction(ctx, map[string]interface{}{"order_id": payload.OrderID})
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
		updates["paid_at"] = time.Now()
	}

	if err := s.repo.UpdateDonationProgramTransaction(ctx, payload.OrderID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":      "donation_program_transaction.service",
			"transaction_id": transaction.ID,
			"order_id":       payload.OrderID,
		}).WithError(err).Error("failed to update transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui transaksi", nil, nil)
	}

	if isSettled {
		prayers, err := s.prayerRepo.FindOnePrayer(ctx, map[string]interface{}{
			"donation_program_transaction_id": transaction.ID,
		})
		if err == nil {
			prayers.IsPublished = true
			if err := s.prayerRepo.UpdatePrayer(ctx, prayers); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":      "donation_program_transaction.service",
					"transaction_id": transaction.ID,
					"order_id":       payload.OrderID,
					"prayer_id":      prayers.ID,
				}).WithError(err).Warn("failed to publish prayer after settlement")
			}
		}
	}

	if isSettled {
		now := time.Now()
		if err := s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
			ID:              uuid.New().String(),
			FundType:        finance_record.FundTypeDonation,
			FundID:          transaction.DonationProgramID.String(),
			SourceType:      finance_record.SourceTypeTransaction,
			SourceID:        transaction.ID.String(),
			Amount:          transaction.GrossAmount,
			TransactionDate: now,
			CreatedAt:       now,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":      "donation_program_transaction.service",
				"transaction_id": transaction.ID,
				"order_id":       payload.OrderID,
			}).WithError(err).Warn("failed to create finance record after settlement")
		}
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_transaction.service",
			"transaction_id":      transaction.ID,
			"order_id":            payload.OrderID,
			"donation_program_id": transaction.DonationProgramID,
			"amount":              transaction.GrossAmount,
		}).Info("transaction settled")
	}

	return pkg.NewResponse(http.StatusOK, "Notifikasi berhasil ditangani", nil, nil)
}

func (s *service) GetMyDonationProgramTransactionList(ctx context.Context, accountID string, params DonationProgramTransactionQueryParams) pkg.Response {
	return s.GetDonationProgramTransactionList(ctx, accountID, "", params)
}

func (s *service) GetMyDonationProgramTransactionByID(ctx context.Context, donationProgramTransactionID string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(donationProgramTransactionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID transaksi tidak valid"}, nil)
	}

	transaction, err := s.repo.FindOneDonationProgramTransaction(ctx, map[string]interface{}{"id": donationProgramTransactionID, "account_id": accountID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Transaksi tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":      "donation_program_transaction.service",
			"transaction_id": donationProgramTransactionID,
			"account_id":     accountID,
		}).WithError(err).Error("failed to fetch transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data transaksi", nil, nil)
	}

	if transaction.AccountID == nil || transaction.AccountID.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Akses ditolak", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, transaction.toDonationProgramTransactionResponse())
}
