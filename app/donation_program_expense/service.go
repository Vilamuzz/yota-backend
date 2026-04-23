package donation_program_expense

import (
	"context"
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
	GetDonationProgramExpenseList(ctx context.Context, donationProgramID string, params DonationProgramExpenseQueryParams) pkg.Response
	GetDonationProgramExpenseByID(ctx context.Context, donationProgramExpenseID string) pkg.Response
	CreateDonationProgramExpense(ctx context.Context, donationProgramID string, payload *DonationProgramExpenseRequest) pkg.Response
	DeleteDonationProgramExpense(ctx context.Context, donationProgramExpenseID string) pkg.Response
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

func (s *service) GetDonationProgramExpenseList(ctx context.Context, donationProgramID string, params DonationProgramExpenseQueryParams) pkg.Response {
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
	if donationProgramID != "" {
		options["donation_program_id"] = donationProgramID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = params.PrevCursor
	}

	expenses, err := s.repo.FindAllDonationProgramExpenses(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation_program_expense.service",
		}).WithError(err).Error("failed to fetch expenses")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch expenses", nil, nil)
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

	return pkg.NewResponse(http.StatusOK, "Expenses found successfully", nil, toDonationProgramExpenseListResponse(expenses, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetDonationProgramExpenseByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	expense, err := s.repo.FindOneDonationProgramExpense(ctx, map[string]interface{}{"id": id})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "donation_program_expense.service",
			"expense_id": id,
		}).WithError(err).Error("failed to fetch expense")
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Expense found successfully", nil, expense.toDonationProgramExpenseDetailResponse())
}

func (s *service) CreateDonationProgramExpense(ctx context.Context, donationProgramID string, payload *DonationProgramExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if donationProgramID == "" {
		errValidation["donation_program_id"] = "Donation Program ID is required"
	} else if _, err := uuid.Parse(donationProgramID); err != nil {
		errValidation["donation_program_id"] = "Invalid donation program ID format"
	}
	if payload.Title == "" {
		errValidation["title"] = "Title is required"
	}
	if payload.Amount <= 0 {
		errValidation["amount"] = "Amount must be greater than 0"
	}
	if payload.ExpenseDate.IsZero() {
		errValidation["expense_date"] = "Expense date is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var proofFileURL string
	if payload.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.ProofFile, "donation-expenses")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload proof file", nil, nil)
		}
		proofFileURL = uploadedURL
	}

	now := time.Now()
	expense := &DonationProgramExpense{
		ID:                uuid.New(),
		DonationProgramID: uuid.MustParse(donationProgramID),
		Title:             payload.Title,
		Amount:            payload.Amount,
		ExpenseDate:       payload.ExpenseDate,
		Note:              payload.Note,
		ProofFile:         proofFileURL,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.CreateDonationProgramExpense(ctx, expense); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "donation_program_expense.service",
			"expense_id": expense.ID,
		}).WithError(err).Error("failed to create expense")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create expense", nil, nil)
	}

	// Auto-create finance record (outflow)
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

	s.logService.CreateLog(ctx, nil, "CREATE", "donation_program_expense", expense.ID.String(), nil, expense.toDonationProgramExpenseDetailResponse())

	return pkg.NewResponse(http.StatusCreated, "Expense created successfully", nil, nil)
}

func (s *service) DeleteDonationProgramExpense(ctx context.Context, donationProgramExpenseID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(donationProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	expense, err := s.repo.FindOneDonationProgramExpense(ctx, map[string]interface{}{"id": donationProgramExpenseID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	if err := s.repo.DeleteDonationProgramExpense(ctx, donationProgramExpenseID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete expense", nil, nil)
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

	// Auto-delete finance record (outflow)
	_ = s.financeRepo.Delete(ctx, donationProgramExpenseID)

	s.logService.CreateLog(ctx, nil, "DELETE", "donation_program_expense", donationProgramExpenseID, expense.toDonationProgramExpenseDetailResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Expense deleted successfully", nil, nil)
}
