package donation_expense

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
)

type Service interface {
	CreateExpense(ctx context.Context, req *CreateExpenseRequest) pkg.Response
	UpdateExpense(ctx context.Context, req *UpdateExpenseRequest) pkg.Response
	DeleteExpense(ctx context.Context, id string) pkg.Response
	GetExpenseByID(ctx context.Context, id string) pkg.Response
	ListExpenses(ctx context.Context, req *QueryParams) pkg.Response
}

type service struct {
	repo        Repository
	financeRepo finance_record.Repository
	s3Client    s3_pkg.Client
	timeout     time.Duration
}

func NewService(repo Repository, financeRepo finance_record.Repository, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:        repo,
		financeRepo: financeRepo,
		s3Client:    s3Client,
		timeout:     timeout,
	}
}

func (s *service) CreateExpense(ctx context.Context, req *CreateExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	if req.DonationID == "" {
		errValidation["donation_id"] = "Donation ID is required"
	} else if _, err := uuid.Parse(req.DonationID); err != nil {
		errValidation["donation_id"] = "Invalid donation ID format"
	}

	if req.Title == "" {
		errValidation["title"] = "Title is required"
	}

	if req.Amount <= 0 {
		errValidation["amount"] = "Amount must be greater than 0"
	}

	if req.Date.IsZero() {
		errValidation["date"] = "Date is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var proofFileURL string
	if req.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.ProofFile, "donation-expenses")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload proof file", nil, nil)
		}
		proofFileURL = uploadedURL
	}

	now := time.Now()
	expense := &DonationExpense{
		ID:         uuid.New().String(),
		DonationID: req.DonationID,
		Title:      req.Title,
		Amount:     req.Amount,
		Date:       req.Date,
		Note:       req.Note,
		ProofFile:  proofFileURL,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, expense); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create expense", nil, nil)
	}

	// Auto-create finance record (outflow)
	_ = s.financeRepo.Create(ctx, &finance_record.FinanceRecord{
		ID:              uuid.New().String(),
		FundType:        finance_record.FundTypeDonation,
		FundID:          expense.DonationID,
		SourceType:      finance_record.SourceTypeExpense,
		SourceID:        expense.ID,
		Amount:          expense.Amount,
		TransactionDate: expense.Date,
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	return pkg.NewResponse(http.StatusCreated, "Expense created successfully", nil, expense)
}

func (s *service) UpdateExpense(ctx context.Context, req *UpdateExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(req.ID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	existing, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	errValidation := make(map[string]string)
	if req.Amount < 0 {
		errValidation["amount"] = "Amount must not be negative"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Amount > 0 {
		existing.Amount = req.Amount
	}
	if !req.Date.IsZero() {
		existing.Date = req.Date
	}
	if req.Note != "" {
		existing.Note = req.Note
	}

	if req.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.ProofFile, "donation-expenses")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload proof file", nil, nil)
		}
		existing.ProofFile = uploadedURL
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update expense", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Expense updated successfully", nil, existing)
}

func (s *service) DeleteExpense(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete expense", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Expense deleted successfully", nil, nil)
}

func (s *service) GetExpenseByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	expense, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Expense found successfully", nil, expense)
}

func (s *service) ListExpenses(ctx context.Context, req *QueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if req.Limit <= 0 {
		req.Limit = 10
	}

	usingPrevCursor := req.PrevCursor != ""

	expenses, err := s.repo.FindAll(ctx, *req)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch expenses", nil, nil)
	}

	hasMore := len(expenses) > req.Limit
	if hasMore {
		expenses = expenses[:req.Limit]
	}

	// Reverse ASC → DESC when navigating backwards
	if usingPrevCursor {
		for i, j := 0, len(expenses)-1; i < j; i, j = i+1, j-1 {
			expenses[i], expenses[j] = expenses[j], expenses[i]
		}
	}

	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && req.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && req.NextCursor != "")

	var nextCursor, prevCursor string
	if hasNext && len(expenses) > 0 {
		last := expenses[len(expenses)-1]
		nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
	}
	if hasPrev && len(expenses) > 0 {
		first := expenses[0]
		prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Expenses found successfully", nil, map[string]interface{}{
		"expenses": expenses,
		"pagination": pkg.CursorPagination{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			Limit:      req.Limit,
		},
	})
}
