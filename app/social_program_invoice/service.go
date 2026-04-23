package social_program_invoice

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetSocialProgramInvoiceList(ctx context.Context, params SocialProgramInvoiceQueryParams) pkg.Response
	GetSocialProgramInvoiceByID(ctx context.Context, id string) pkg.Response
	CreateSocialProgramInvoice(ctx context.Context, payload SocialProgramInvoiceRequest) pkg.Response
	UpdateSocialProgramInvoice(ctx context.Context, id string, payload SocialProgramInvoiceRequest) pkg.Response
	DeleteSocialProgramInvoice(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *service) GetSocialProgramInvoiceList(ctx context.Context, params SocialProgramInvoiceQueryParams) pkg.Response {
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
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = params.PrevCursor
	}
	if params.SubscriptionID != "" {
		options["subscription_id"] = params.SubscriptionID
	}
	if params.Status != "" {
		options["status"] = params.Status
	}

	invoices, err := s.repo.FindAllSocialProgramInvoices(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_invoice.service",
		}).WithError(err).Error("failed to fetch invoices")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch invoices", nil, nil)
	}

	hasMore := len(invoices) > params.Limit
	if hasMore {
		invoices = invoices[:params.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(invoices)-1; i < j; i, j = i+1, j-1 {
			invoices[i], invoices[j] = invoices[j], invoices[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(invoices) > 0 {
		first := invoices[0]
		last := invoices[len(invoices)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Invoices found successfully", nil, toSocialProgramInvoiceListResponse(invoices, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetSocialProgramInvoiceByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid invoice ID format"}, nil)
	}

	invoice, err := s.repo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Invoice not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_invoice.service",
			"invoice_id": id,
		}).WithError(err).Error("failed to fetch invoice")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch invoice", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Invoice found successfully", nil, invoice.toSocialProgramInvoiceResponse())
}

func (s *service) CreateSocialProgramInvoice(ctx context.Context, payload SocialProgramInvoiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(payload.SubscriptionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"subscription_id": "Invalid subscription ID format"}, nil)
	}

	now := time.Now()
	invoice := &SocialProgramInvoice{
		ID:             uuid.New(),
		SubscriptionID: uuid.MustParse(payload.SubscriptionID),
		BillingPeriod:  payload.BillingPeriod,
		Amount:         payload.Amount,
		Status:         payload.Status,
		DueDate:        payload.DueDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateSocialProgramInvoice(ctx, invoice); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_invoice.service",
		}).WithError(err).Error("failed to create invoice")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create invoice", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Invoice created successfully", nil, invoice.toSocialProgramInvoiceResponse())
}

func (s *service) UpdateSocialProgramInvoice(ctx context.Context, id string, payload SocialProgramInvoiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid invoice ID format"}, nil)
	}

	if _, err := uuid.Parse(payload.SubscriptionID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"subscription_id": "Invalid subscription ID format"}, nil)
	}

	_, err := s.repo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Invoice not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch invoice", nil, nil)
	}

	updates := map[string]interface{}{
		"subscription_id": uuid.MustParse(payload.SubscriptionID),
		"billing_period":  payload.BillingPeriod,
		"amount":          payload.Amount,
		"status":          payload.Status,
		"due_date":        payload.DueDate,
		"updated_at":      time.Now(),
	}

	if err := s.repo.UpdateSocialProgramInvoice(ctx, id, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_invoice.service",
			"invoice_id": id,
		}).WithError(err).Error("failed to update invoice")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update invoice", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Invoice updated successfully", nil, nil)
}

func (s *service) DeleteSocialProgramInvoice(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid invoice ID format"}, nil)
	}

	_, err := s.repo.FindOneSocialProgramInvoice(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Invoice not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch invoice", nil, nil)
	}

	if err := s.repo.DeleteSocialProgramInvoice(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_invoice.service",
			"invoice_id": id,
		}).WithError(err).Error("failed to delete invoice")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete invoice", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Invoice deleted successfully", nil, nil)
}
