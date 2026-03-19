package donation_expense

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
)

type Service interface {
	CreateExpense(ctx context.Context, payload *CreateExpenseRequest) pkg.Response
	UpdateExpense(ctx context.Context, payload *UpdateExpenseRequest) pkg.Response
	DeleteExpense(ctx context.Context, id string) pkg.Response
	GetExpenseByID(ctx context.Context, id string) pkg.Response
	ListExpenses(ctx context.Context, queryParams DonationExpenseQueryParams) pkg.Response
}

type service struct {
	repo         Repository
	financeRepo  finance_record.Repository
	donationRepo donation.Repository
	s3Client     s3_pkg.Client
	timeout      time.Duration
}

func NewService(repo Repository, financeRepo finance_record.Repository, donationRepo donation.Repository, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:         repo,
		financeRepo:  financeRepo,
		donationRepo: donationRepo,
		s3Client:     s3Client,
		timeout:      timeout,
	}
}

func (s *service) ListExpenses(ctx context.Context, queryParams DonationExpenseQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit <= 0 {
		queryParams.Limit = 10
	}

	usingPrevCursor := queryParams.PrevCursor != ""

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.DonationID != "" {
		options["donation_id"] = queryParams.DonationID
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	expenses, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch expenses", nil, nil)
	}

	hasMore := len(expenses) > queryParams.Limit
	if hasMore {
		expenses = expenses[:queryParams.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(expenses)-1; i < j; i, j = i+1, j-1 {
			expenses[i], expenses[j] = expenses[j], expenses[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && queryParams.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && queryParams.NextCursor != "")

	if len(expenses) > 0 {
		first := expenses[0]
		last := expenses[len(expenses)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	return pkg.NewResponse(http.StatusOK, "Expenses found successfully", nil, toDonationExpenseListResponse(expenses, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
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

	return pkg.NewResponse(http.StatusOK, "Expense found successfully", nil, expense.toDonationExpenseDetailResponse())
}

func (s *service) CreateExpense(ctx context.Context, payload *CreateExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.DonationID == "" {
		errValidation["donation_id"] = "Donation ID is required"
	} else if _, err := uuid.Parse(payload.DonationID); err != nil {
		errValidation["donation_id"] = "Invalid donation ID format"
	}
	if payload.Title == "" {
		errValidation["title"] = "Title is required"
	}
	if payload.Amount <= 0 {
		errValidation["amount"] = "Amount must be greater than 0"
	}
	if payload.Date.IsZero() {
		errValidation["date"] = "Date is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Validate amount to not exceed collected fund
	d, err := s.donationRepo.FindOne(ctx, map[string]interface{}{"id": payload.DonationID})
	if err != nil {
		errValidation["donation_id"] = "Donation not found"
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	totalExpenses, err := s.repo.GetTotalExpenseByDonationID(ctx, payload.DonationID)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to calculate total expenses", nil, nil)
	}

	if totalExpenses+payload.Amount > d.CollectedFund {
		errValidation["amount"] = "Expense amount exceeds total collected funds for this donation"
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
	expense := &DonationExpense{
		ID:         uuid.New().String(),
		DonationID: payload.DonationID,
		Title:      payload.Title,
		Amount:     payload.Amount,
		Date:       payload.Date,
		Note:       payload.Note,
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

	return pkg.NewResponse(http.StatusCreated, "Expense created successfully", nil, nil)
}

func (s *service) UpdateExpense(ctx context.Context, payload *UpdateExpenseRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(payload.ID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid expense ID format"}, nil)
	}

	existing, err := s.repo.FindByID(ctx, payload.ID)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Expense not found", nil, nil)
	}

	updateData := make(map[string]interface{})
	errValidation := make(map[string]string)
	if payload.Amount < 0 {
		errValidation["amount"] = "Amount must not be negative"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Validate against collected fund if amount is updating
	if payload.Amount > 0 && payload.Amount != existing.Amount {
		d, err := s.donationRepo.FindOne(ctx, map[string]interface{}{"id": existing.DonationID})
		if err != nil {
			errValidation["donation_id"] = "Donation not found"
			return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
		}

		totalExpenses, err := s.repo.GetTotalExpenseByDonationID(ctx, existing.DonationID)
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to calculate total expenses", nil, nil)
		}

		// Check if the old amount being replaced + new amount exceeds collected fund
		if (totalExpenses-existing.Amount)+payload.Amount > d.CollectedFund {
			errValidation["amount"] = "Expense amount exceeds total collected funds for this donation"
			return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
		}
	}

	if payload.Title != "" {
		updateData["title"] = payload.Title
	}
	if payload.Amount > 0 {
		updateData["amount"] = payload.Amount
	}
	if !payload.Date.IsZero() {
		updateData["date"] = payload.Date
	}
	if payload.Note != "" {
		updateData["note"] = payload.Note
	}

	if payload.ProofFile != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.ProofFile, "donation-expenses")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload proof file", nil, nil)
		}
		updateData["proof_file"] = uploadedURL
	}

	if err := s.repo.Update(ctx, existing.ID, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update expense", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Expense updated successfully", nil, nil)
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
