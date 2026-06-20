package donation_program_expense

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetDonationProgramExpenseList(ctx context.Context, slug string, params DonationProgramExpenseQueryParams) pkg.Response
	GetAdminDonationProgramExpenseList(ctx context.Context, donationProgramID string, params DonationProgramExpenseQueryParams) pkg.Response
	GetDonationProgramExpenseByID(ctx context.Context, donationProgramExpenseID string) pkg.Response
	CreateDonationProgramExpense(ctx context.Context, accountID, donationProgramID string, payload *DonationProgramExpenseRequest) pkg.Response
	DeleteDonationProgramExpense(ctx context.Context, accountID, donationProgramExpenseID string) pkg.Response
	ExportDonationProgramExpenseCSV(ctx context.Context, donationProgramIdentifier string, params DonationProgramExpenseQueryParams) ([]byte, string, error)
	GetDonationExpenseMonthlyExpense(ctx context.Context, donationProgramID string, params MonthlyExpenseQueryParams) pkg.Response
}

type service struct {
	repo         Repository
	financeRepo  finance_record.Repository
	donationRepo donation_program.Repository
	s3Client     s3_pkg.Client
	logService   app_log.Service
	timeout      time.Duration
}

func NewService(repo Repository, financeRepo finance_record.Repository, donationRepo donation_program.Repository, s3Client s3_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:         repo,
		financeRepo:  financeRepo,
		donationRepo: donationRepo,
		s3Client:     s3Client,
		logService:   logService,
		timeout:      timeout,
	}
}

func (s *service) GetDonationProgramExpenseList(ctx context.Context, slug string, params DonationProgramExpenseQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": slug})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Program donasi tidak ditemukan", nil, nil)
	}

	return s.GetAdminDonationProgramExpenseList(ctx, program.ID.String(), params)
}

func (s *service) GetAdminDonationProgramExpenseList(ctx context.Context, donationProgramID string, params DonationProgramExpenseQueryParams) pkg.Response {
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
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
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

	expenses, err := s.repo.FindAllDonationProgramExpenses(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation_program_expense.service",
		}).WithError(err).Error("failed to fetch expenses")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pengeluaran", nil, nil)
	}

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(expenses) > params.Limit
		hasNext = true
		if len(expenses) > params.Limit {
			expenses = expenses[:params.Limit]
		}
		for i, j := 0, len(expenses)-1; i < j; i, j = i+1, j-1 {
			expenses[i], expenses[j] = expenses[j], expenses[i]
		}
	} else {
		hasNext = len(expenses) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			expenses = expenses[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
	if len(expenses) > 0 {
		first := expenses[0]
		last := expenses[len(expenses)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toDonationProgramExpenseListResponse(expenses, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetDonationProgramExpenseByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID pengeluaran tidak valid"}, nil)
	}

	expense, err := s.repo.FindOneDonationProgramExpense(ctx, map[string]interface{}{"id": id})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "donation_program_expense.service",
			"expense_id": id,
		}).WithError(err).Error("failed to fetch expense")
		return pkg.NewResponse(http.StatusNotFound, "Pengeluaran tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, expense.toDonationProgramExpenseDetailResponse())
}

func (s *service) CreateDonationProgramExpense(ctx context.Context, accountID, donationProgramID string, payload *DonationProgramExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if donationProgramID == "" {
		errValidation["donationProgramId"] = "ID Program Donasi wajib diisi"
	} else if err := uuid.Validate(donationProgramID); err != nil {
		errValidation["donationProgramId"] = "Format ID program donasi tidak valid"
	}
	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	}
	if payload.Amount <= 0 {
		errValidation["amount"] = "Jumlah harus lebih besar dari 0"
	}
	if payload.ExpenseDate == "" {
		errValidation["expenseDate"] = "Tanggal pengeluaran wajib diisi"
	} else {
		if _, err := time.Parse("2006-01-02", payload.ExpenseDate); err != nil {
			errValidation["expenseDate"] = "Format tanggal tidak valid (gunakan YYYY-MM-DD)"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	donationProgram, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"id": donationProgramID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Program donasi tidak ditemukan", nil, nil)
	}

	availableFund := donationProgram.CollectedFund - donationProgram.TotalExpense
	if payload.Amount > availableFund {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"amount": "Jumlah pengeluaran melebihi dana yang tersedia"}, nil)
	}

	var proofFileURL string
	if payload.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.ProofFile, "donation-expenses")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "donation_program_expense.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload proof file")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file bukti", nil, nil)
		}
		proofFileURL = uploadedURL
	}

	now := time.Now()
	expenseDate, _ := time.Parse("2006-01-02", payload.ExpenseDate)
	expense := &DonationProgramExpense{
		ID:                uuid.New(),
		DonationProgramID: uuid.MustParse(donationProgramID),
		Title:             payload.Title,
		Amount:            payload.Amount,
		ExpenseDate:       expenseDate,
		Note:              payload.Note,
		ProofFile:         proofFileURL,
		CreatedBy:         uuid.MustParse(accountID),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.CreateDonationProgramExpense(ctx, expense); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "donation_program_expense.service",
			"expense_id": expense.ID,
		}).WithError(err).Error("failed to create expense")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat pengeluaran", nil, nil)
	}

	_ = s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeDonation,
		FundID:          expense.DonationProgramID.String(),
		SourceType:      finance_record.SourceTypeExpense,
		SourceID:        expense.ID.String(),
		Amount:          expense.Amount,
		TransactionDate: expense.ExpenseDate,
		CreatedAt:       now,
	})

	s.logService.CreateLog(ctx, &accountID, "CREATE", "donation_program_expense", expense.ID.String(), nil, expense.toDonationProgramExpenseDetailResponse())

	return pkg.NewResponse(http.StatusCreated, "Pengeluaran berhasil dibuat", nil, nil)
}

func (s *service) ExportDonationProgramExpenseCSV(ctx context.Context, donationProgramIdentifier string, params DonationProgramExpenseQueryParams) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	filter := map[string]interface{}{}
	if err := uuid.Validate(donationProgramIdentifier); err == nil {
		filter["id"] = donationProgramIdentifier
	} else {
		filter["slug"] = donationProgramIdentifier
	}

	program, err := s.donationRepo.FindOneDonationProgram(ctx, filter)
	if err != nil {
		return nil, "", fmt.Errorf("program donasi tidak ditemukan")
	}
	donationProgramID := program.ID.String()

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

	expenses, err := s.repo.FindAllDonationProgramExpensesForExport(ctx, donationProgramID, params)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_expense.service",
			"donation_program_id": donationProgramID,
		}).WithError(err).Error("failed to fetch expenses for export")
		return nil, "", fmt.Errorf("gagal mengambil data pengeluaran")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{"No", "Judul", "Jumlah (Rp)", "Tanggal Pengeluaran", "Catatan", "Dibuat Pada"}
	if err := w.Write(header); err != nil {
		return nil, "", fmt.Errorf("gagal menulis header CSV")
	}

	for i, expense := range expenses {
		row := []string{
			fmt.Sprintf("%d", i+1),
			expense.Title,
			fmt.Sprintf("%.2f", expense.Amount),
			expense.ExpenseDate.Format("2006-01-02"),
			expense.Note,
			expense.CreatedAt.Format("2006-01-02 15:04:05"),
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
	filename := fmt.Sprintf("donation_program_expenses_%s_%s_%s.csv", donationProgramID, periodPart, time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

func (s *service) DeleteDonationProgramExpense(ctx context.Context, accountID, donationProgramExpenseID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(donationProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID pengeluaran tidak valid"}, nil)
	}

	expense, err := s.repo.FindOneDonationProgramExpense(ctx, map[string]interface{}{"id": donationProgramExpenseID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Pengeluaran tidak ditemukan", nil, nil)
	}

	if err := s.repo.DeleteDonationProgramExpense(ctx, donationProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus pengeluaran", nil, nil)
	}

	if expense.ProofFile != "" {
		imageObjectName := s3_pkg.ExtractObjectNameFromURL(expense.ProofFile)
		if err := s.s3Client.DeleteFile(ctx, imageObjectName); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "donation_program_expense.service",
				"expense_id": donationProgramExpenseID,
			}).WithError(err).Error("failed to delete proof file from S3")
		}
	}

	s.logService.CreateLog(ctx, &accountID, "DELETE", "donation_program_expense", donationProgramExpenseID, expense.toDonationProgramExpenseDetailResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Pengeluaran berhasil dihapus", nil, nil)
}

func (s *service) GetDonationExpenseMonthlyExpense(ctx context.Context, donationProgramID string, params MonthlyExpenseQueryParams) pkg.Response {
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

	expenseRecord, err := s.repo.GetMonthlyExpenseByProgram(ctx, donationProgramID, yearVal)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":           "donation_program_expense.service",
			"donation_program_id": donationProgramID,
		}).WithError(err).Error("failed to get monthly expense")

		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pengeluaran bulanan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, expenseRecord)
}
