package donation_program_transaction

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/csv"
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
	GetDonationTransactionMonthlyIncome(ctx context.Context, donationProgramID string, params MonthlyIncomeQueryParams) pkg.Response
	ExportDonationProgramTransactionCSV(ctx context.Context, donationProgramID string, params DonationProgramTransactionQueryParams) ([]byte, string, error)

	HandleNotification(ctx context.Context, payload payment_pkg.MidtransNotificationRequest) pkg.Response

	GetMyDonationProgramTransactionList(ctx context.Context, accountID string, params DonationProgramTransactionQueryParams) pkg.Response
	GetMyDonationProgramTransactionByID(ctx context.Context, donationProgramTransactionID, accountID string) pkg.Response
	GetPublicDonationProgramTransactionList(ctx context.Context, slug string, params DonationProgramTransactionQueryParams) pkg.Response
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

	errValidation := make(map[string]string)
	if params.StartDate != "" {
		if _, err := time.Parse("2006-01-02", params.StartDate); err != nil {
			errValidation["startDate"] = "Format tanggal tidak valid (gunakan YYYY-MM-DD)"
		}
	}
	if params.EndDate != "" {
		if _, err := time.Parse("2006-01-02", params.EndDate); err != nil {
			errValidation["endDate"] = "Format tanggal tidak valid (gunakan YYYY-MM-DD)"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if donationProgramID != "" {
		options["donation_program_id"] = donationProgramID
	}
	if accountID != "" {
		options["account_id"] = accountID
	}
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}
	if params.StartDate != "" {
		options["start_date"] = params.StartDate
	}
	if params.EndDate != "" {
		options["end_date"] = params.EndDate
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

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(transactions) > params.Limit
		hasNext = true
		if len(transactions) > params.Limit {
			transactions = transactions[:params.Limit]
		}
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
		}
	} else {
		hasNext = len(transactions) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			transactions = transactions[:params.Limit]
		}
	}

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

	if err := uuid.Validate(donationProgramTransactionID); err != nil {
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

	var donationProg *donation_program.DonationProgram
	if donationProgramID == "" {
		errValidation["donationProgramId"] = "ID Program Donasi wajib diisi"
	} else {
		prog, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"id": donationProgramID})
		if err != nil {
			errValidation["donationProgramId"] = "Program Donasi tidak ditemukan"
		} else if prog.Status == donation_program.StatusExpired || prog.Status == donation_program.StatusCompleted {
			errValidation["donationProgramId"] = "Program donasi sudah kedaluwarsa atau selesai"
		} else if prog.Status != donation_program.StatusActive {
			errValidation["donationProgramId"] = "Program Donasi tidak aktif"
		} else {
			donationProg = prog
		}
	}

	if payload.GrossAmount <= 0 {
		errValidation["grossAmount"] = "Jumlah kotor harus lebih besar dari 0"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	donorName := "Hamba Allah"
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

	transaction.DonationProgram = donationProg
	s.logService.CreateLog(ctx, &accountID, "CREATE", "donation_program_transaction", transaction.ID.String(), nil, transaction.toDonationProgramTransactionResponse())

	return pkg.NewResponse(http.StatusCreated, "Transaksi offline berhasil dibuat", nil, transaction.toDonationProgramTransactionResponse())
}

func (s *service) CreateDonationProgramTransaction(ctx context.Context, accountID string, donationSlug string, payload CreateDonationProgramTransactionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	var donationProgramID string

	var donationProg *donation_program.DonationProgram
	if donationSlug == "" {
		errValidation["donationProgramSlug"] = "Slug Program Donasi wajib diisi"
	} else {
		program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": donationSlug})
		if err != nil {
			errValidation["donationProgramSlug"] = "Program Donasi tidak ditemukan"
		} else if program.Status == donation_program.StatusExpired || program.Status == donation_program.StatusCompleted {
			errValidation["donationProgramSlug"] = "Program donasi sudah kedaluwarsa atau selesai"
		} else if program.Status != donation_program.StatusActive {
			errValidation["donationProgramSlug"] = "Program Donasi tidak aktif"
		} else {
			donationProgramID = program.ID.String()
			donationProg = program
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

	transaction.DonationProgram = donationProg
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

	if err := uuid.Validate(donationProgramTransactionID); err != nil {
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

func (s *service) GetPublicDonationProgramTransactionList(ctx context.Context, slug string, params DonationProgramTransactionQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": slug})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program Donasi tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program donasi", nil, nil)
	}

	// For public route, only show settled transactions
	params.Status = string(TransactionStatusSettlement)

	return s.GetDonationProgramTransactionList(ctx, "", program.ID.String(), params)
}

func (s *service) GetDonationTransactionMonthlyIncome(ctx context.Context, donationProgramID string, params MonthlyIncomeQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(donationProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program donasi tidak valid"}, nil)
	}

	yearVal := time.Now().Year()
	if params.Year != "" {
		var parseYear int
		if _, err := fmt.Sscanf(params.Year, "%d", &parseYear); err == nil && parseYear > 0 {
			yearVal = parseYear
		}
	}

	incomeRecord, err := s.repo.GetMonthlyIncomeByProgram(ctx, donationProgramID, yearVal)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_transaction.service",
			"donation_program_id": donationProgramID,
		}).WithError(err).Error("failed to get monthly income")

		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pendapatan bulanan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, incomeRecord)
}

func (s *service) ExportDonationProgramTransactionCSV(ctx context.Context, donationProgramID string, params DonationProgramTransactionQueryParams) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"id": donationProgramID})
	if err != nil {
		return nil, "", fmt.Errorf("program donasi tidak ditemukan")
	}
	donationProgramID = program.ID.String()

	if params.StartDate != "" {
		if _, err := time.Parse("2006-01-02", params.StartDate); err != nil {
			return nil, "", fmt.Errorf("format start_date tidak valid (gunakan YYYY-MM-DD)")
		}
	}
	if params.EndDate != "" {
		if _, err := time.Parse("2006-01-02", params.EndDate); err != nil {
			return nil, "", fmt.Errorf("format end_date tidak valid (gunakan YYYY-MM-DD)")
		}
	}

	transactions, err := s.repo.FindAllDonationProgramTransactionsForExport(ctx, donationProgramID, params)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_transaction.service",
			"donation_program_id": donationProgramID,
		}).WithError(err).Error("failed to fetch transactions for export")
		return nil, "", fmt.Errorf("gagal mengambil data transaksi")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{"No", "Order ID", "Nama Donatur", "Email Donatur", "Tipe Transaksi", "Jumlah (Rp)", "Metode Pembayaran", "Status Transaksi", "Tanggal Bayar", "Tanggal Dibuat"}
	if err := w.Write(header); err != nil {
		return nil, "", fmt.Errorf("gagal menulis header CSV")
	}

	for i, tx := range transactions {
		typeStr := "Offline"
		if tx.IsOnline {
			typeStr = "Online"
		}

		paidAtStr := "-"
		if tx.PaidAt != nil {
			paidAtStr = tx.PaidAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", i+1),
			tx.OrderID,
			tx.DonorName,
			tx.DonorEmail,
			typeStr,
			fmt.Sprintf("%.2f", tx.GrossAmount),
			tx.Provider,
			tx.TransactionStatus,
			paidAtStr,
			tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := w.Write(row); err != nil {
			return nil, "", fmt.Errorf("gagal menulis baris CSV")
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", fmt.Errorf("gagal menyelesaikan penulisan CSV")
	}

	periodPart := "all"
	if params.StartDate != "" && params.EndDate != "" {
		periodPart = params.StartDate + "_to_" + params.EndDate
	} else if params.StartDate != "" {
		periodPart = "from_" + params.StartDate
	} else if params.EndDate != "" {
		periodPart = "until_" + params.EndDate
	}
	filename := fmt.Sprintf("donation_program_transactions_%s_%s_%s.csv", donationProgramID, periodPart, time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}
