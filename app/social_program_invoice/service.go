package social_program_invoice

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetSocialProgramInvoiceList(ctx context.Context, params SocialProgramInvoiceQueryParams) pkg.Response
	GetSocialProgramInvoiceByID(ctx context.Context, id string) pkg.Response
	GenerateMonthlyInvoices(ctx context.Context) error
}

type service struct {
	repo             Repository
	subscriptionRepo social_program_subscription.Repository
	timeout          time.Duration
}

func NewService(repo Repository, subscriptionRepo social_program_subscription.Repository, timeout time.Duration) Service {
	return &service{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		timeout:          timeout,
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

func (s *service) GenerateMonthlyInvoices(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5) // generous timeout for batch job
	defer cancel()

	now := time.Now()
	
	// Determine the target billing day
	// If today is the last day of the month, we also need to generate invoices for
	// programs whose billing day is > today (e.g. billing day 31, but today is 30th)
	// For simplicity right now, we will match exact days, but adjust for end of month.
	
	currentDay := now.Day()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	
	var targetDays []int
	targetDays = append(targetDays, currentDay)
	
	// If today is the last day of the month, include all days greater than today
	if currentDay == daysInMonth {
		for i := currentDay + 1; i <= 31; i++ {
			targetDays = append(targetDays, i)
		}
	}

	for _, day := range targetDays {
		subscriptions, err := s.subscriptionRepo.FindSubscriptionsDueForBilling(ctx, day)
		if err != nil {
			logrus.WithError(err).Error("Failed to fetch subscriptions for billing")
			continue
		}

		for _, sub := range subscriptions {
			// Calculate billing period (start of the current month)
			billingPeriod := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			dueDate := time.Date(now.Year(), now.Month(), day, 23, 59, 59, 0, now.Location())
			
			// If we're generating on the last day but the billing day is technically later,
			// just use the end of this month as due date
			if dueDate.Month() != now.Month() {
				dueDate = time.Date(now.Year(), now.Month(), daysInMonth, 23, 59, 59, 0, now.Location())
			}

			// Check if invoice already exists for this subscription and period
			options := map[string]interface{}{
				"subscription_id": sub.ID.String(),
				"billing_period":  billingPeriod.Format("2006-01-02"),
			}
			existing, _ := s.repo.FindOneSocialProgramInvoice(ctx, options)
			if existing != nil {
				continue // Already billed for this period
			}

			invoice := &SocialProgramInvoice{
				ID:             uuid.New(),
				SubscriptionID: sub.ID,
				BillingPeriod:  billingPeriod,
				MinimumAmount:  sub.SocialProgram.MinimumAmount,
				Status:         InvoiceStatusPending,
				DueDate:        dueDate,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := s.repo.CreateSocialProgramInvoice(ctx, invoice); err != nil {
				logrus.WithFields(logrus.Fields{
					"subscription_id": sub.ID,
					"error":           err,
				}).Error("Failed to create monthly invoice")
			} else {
				logrus.WithFields(logrus.Fields{
					"invoice_id":      invoice.ID,
					"subscription_id": sub.ID,
				}).Info("Monthly invoice generated")
			}
		}
	}
	
	return nil
}
