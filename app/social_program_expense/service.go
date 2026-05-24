package social_program_expense

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/finance_record"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetPublicSocialProgramExpenseList(ctx context.Context, socialProgramSlug string, params SocialProgramExpenseQueryParams) pkg.Response
	GetSocialProgramExpenseList(ctx context.Context, socialProgramID string, params SocialProgramExpenseQueryParams) pkg.Response
	GetSocialProgramExpenseByID(ctx context.Context, socialProgramExpenseID string) pkg.Response
	CreateSocialProgramExpense(ctx context.Context, accountID string, socialProgramID string, payload *SocialProgramExpenseRequest) pkg.Response
	DeleteSocialProgramExpense(ctx context.Context, accountID, socialProgramExpenseID string) pkg.Response
}

type service struct {
	repo              Repository
	financeRepo       finance_record.Repository
	socialProgramRepo social_program.Repository
	s3Client          s3_pkg.Client
	logService        app_log.Service
	timeout           time.Duration
}

func NewService(repo Repository, financeRepo finance_record.Repository, socialProgramRepo social_program.Repository, s3Client s3_pkg.Client, logService app_log.Service, timeout time.Duration) Service {
	return &service{
		repo:              repo,
		financeRepo:       financeRepo,
		socialProgramRepo: socialProgramRepo,
		s3Client:          s3Client,
		logService:        logService,
		timeout:           timeout,
	}
}

func (s *service) GetPublicSocialProgramExpenseList(ctx context.Context, socialProgramSlug string, params SocialProgramExpenseQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	program, err := s.socialProgramRepo.FindOneSocialProgram(ctx, map[string]interface{}{"slug": socialProgramSlug})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
	}

	return s.GetSocialProgramExpenseList(ctx, program.ID.String(), params)
}

func (s *service) GetSocialProgramExpenseList(ctx context.Context, socialProgramID string, params SocialProgramExpenseQueryParams) pkg.Response {
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
		"limit":             params.Limit,
		"social_program_id": socialProgramID,
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = params.PrevCursor
	}

	expenses, err := s.repo.FindAllSocialProgramExpenses(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_expense.service",
		}).WithError(err).Error("failed to fetch expenses")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pengeluaran", nil, nil)
	}

	hasMore := len(expenses) > params.Limit
	if hasMore {
		expenses = expenses[:params.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(expenses)-1; i < j; i, j = i+1, j-1 {
			expenses[i], expenses[j] = expenses[j], expenses[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toSocialProgramExpenseListResponse(expenses, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetSocialProgramExpenseByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID pengeluaran tidak valid"}, nil)
	}

	expense, err := s.repo.FindOneSocialProgramExpense(ctx, map[string]interface{}{"id": id})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_expense.service",
			"expense_id": id,
		}).WithError(err).Error("failed to fetch expense")
		return pkg.NewResponse(http.StatusNotFound, "Pengeluaran tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, expense.toSocialProgramExpenseDetailResponse())
}

func (s *service) CreateSocialProgramExpense(ctx context.Context, accountID string, socialProgramID string, payload *SocialProgramExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if socialProgramID == "" {
		errValidation["socialProgramId"] = "ID Program Sosial wajib diisi"
	} else if err := uuid.Validate(socialProgramID); err != nil {
		errValidation["socialProgramId"] = "Format ID program sosial tidak valid"
	}
	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	}
	if payload.Amount <= 0 {
		errValidation["amount"] = "Jumlah harus lebih besar dari 0"
	}
	if payload.ExpenseDate.IsZero() {
		errValidation["expenseDate"] = "Tanggal pengeluaran wajib diisi"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	socialProgram, err := s.socialProgramRepo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
	}

	availableFund := socialProgram.CollectedFund - socialProgram.TotalExpense
	if payload.Amount > availableFund {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"amount": "Jumlah pengeluaran melebihi dana yang tersedia"}, nil)
	}

	var proofFileURL string
	if payload.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.ProofFile, "social-program-expenses")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "social_program_expense.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload proof file")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file bukti", nil, nil)
		}
		proofFileURL = uploadedURL
	}

	now := time.Now()
	expense := &SocialProgramExpense{
		ID:              uuid.New(),
		SocialProgramID: uuid.MustParse(socialProgramID),
		Title:           payload.Title,
		Amount:          payload.Amount,
		ExpenseDate:     payload.ExpenseDate,
		Note:            payload.Note,
		ProofFile:       proofFileURL,
		CreatedBy:       uuid.MustParse(accountID),
		CreatedAt:       now,
	}

	if err := s.repo.CreateSocialProgramExpense(ctx, expense); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_expense.service",
			"expense_id": expense.ID,
		}).WithError(err).Error("failed to create expense")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat pengeluaran", nil, nil)
	}

	// Auto-create finance record (outflow)
	_ = s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeSocialProgram,
		FundID:          expense.SocialProgramID.String(),
		SourceType:      finance_record.SourceTypeExpense,
		SourceID:        expense.ID.String(),
		Amount:          expense.Amount,
		TransactionDate: expense.ExpenseDate,
		CreatedAt:       now,
	})

	s.logService.CreateLog(ctx, &accountID, "CREATE", "social_program_expense", expense.ID.String(), nil, expense.toSocialProgramExpenseDetailResponse())

	return pkg.NewResponse(http.StatusCreated, "Pengeluaran berhasil dibuat", nil, nil)
}

func (s *service) DeleteSocialProgramExpense(ctx context.Context, accountID, socialProgramExpenseID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID pengeluaran tidak valid"}, nil)
	}

	expense, err := s.repo.FindOneSocialProgramExpense(ctx, map[string]interface{}{"id": socialProgramExpenseID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Pengeluaran tidak ditemukan", nil, nil)
	}

	if err := s.repo.DeleteSocialProgramExpense(ctx, socialProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus pengeluaran", nil, nil)
	}

	if expense.ProofFile != "" {
		imageObjectName := s3_pkg.ExtractObjectNameFromURL(expense.ProofFile)
		if err := s.s3Client.DeleteFile(ctx, imageObjectName); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "social_program_expense.service",
				"expense_id": socialProgramExpenseID,
			}).WithError(err).Error("failed to delete proof file from S3")
		}
	}

	// Auto-delete finance record (outflow)
	_ = s.financeRepo.Delete(ctx, socialProgramExpenseID)

	accountID = expense.CreatedBy.String()
	s.logService.CreateLog(ctx, &accountID, "DELETE", "social_program_expense", socialProgramExpenseID, expense.toSocialProgramExpenseDetailResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Pengeluaran berhasil dihapus", nil, nil)
}
