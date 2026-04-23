package social_program_subscription

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetSocialProgramSubscriptionList(ctx context.Context, params SocialProgramSubscriptionQueryParams) pkg.Response
	GetSocialProgramSubscriptionByID(ctx context.Context, id string) pkg.Response
	CreateSocialProgramSubscription(ctx context.Context, accountID string, req CreateSocialProgramSubscriptionRequest) pkg.Response
	UpdateSocialProgramSubscription(ctx context.Context, id string, req UpdateSocialProgramSubscriptionRequest) pkg.Response
	DeleteSocialProgramSubscription(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo              Repository
	socialProgramRepo social_program.Repository
	timeout           time.Duration
}

func NewService(repo Repository, socialProgramRepo social_program.Repository, timeout time.Duration) Service {
	return &service{
		repo:              repo,
		socialProgramRepo: socialProgramRepo,
		timeout:           timeout,
	}
}

func (s *service) GetSocialProgramSubscriptionList(ctx context.Context, params SocialProgramSubscriptionQueryParams) pkg.Response {
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
	if params.SocialProgramID != "" {
		options["social_program_id"] = params.SocialProgramID
	}
	if params.AccountID != "" {
		options["account_id"] = params.AccountID
	}
	if params.Status != "" {
		options["status"] = params.Status
	}

	subscriptions, err := s.repo.FindAllSocialProgramSubscriptions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to fetch subscriptions")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch subscriptions", nil, nil)
	}

	hasMore := len(subscriptions) > params.Limit
	if hasMore {
		subscriptions = subscriptions[:params.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(subscriptions)-1; i < j; i, j = i+1, j-1 {
			subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(subscriptions) > 0 {
		first := subscriptions[0]
		last := subscriptions[len(subscriptions)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Subscriptions found successfully", nil, toSocialProgramSubscriptionListResponse(subscriptions, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetSocialProgramSubscriptionByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid subscription ID format"}, nil)
	}

	subscription, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Subscription not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to fetch subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch subscription", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Subscription found successfully", nil, subscription.toSocialProgramSubscriptionResponse())
}

func (s *service) CreateSocialProgramSubscription(ctx context.Context, accountID string, req CreateSocialProgramSubscriptionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(req.SocialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"social_program_id": "Invalid social program ID format"}, nil)
	}
	
	// verify social program exists
	if _, err := s.socialProgramRepo.FindOneSocialProgram(ctx, map[string]interface{}{"id": req.SocialProgramID}); err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Social program not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch social program", nil, nil)
	}

	// check if existing active subscription
	existing, _ := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{
		"social_program_id": req.SocialProgramID,
		"account_id":        accountID,
		"status":            string(StatusActive),
	})
	if existing != nil {
		return pkg.NewResponse(http.StatusConflict, "Active subscription already exists for this social program", nil, nil)
	}

	now := time.Now()
	subscription := &SocialProgramSubscription{
		ID:              uuid.New(),
		SocialProgramID: uuid.MustParse(req.SocialProgramID),
		AccountID:       uuid.MustParse(accountID),
		Status:          StatusActive,
		Amount:          req.Amount,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateSocialProgramSubscription(ctx, subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to create subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create subscription", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Subscription created successfully", nil, subscription.toSocialProgramSubscriptionResponse())
}

func (s *service) UpdateSocialProgramSubscription(ctx context.Context, id string, req UpdateSocialProgramSubscriptionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid subscription ID format"}, nil)
	}

	_, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Subscription not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch subscription", nil, nil)
	}

	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateSocialProgramSubscription(ctx, id, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to update subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update subscription", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Subscription updated successfully", nil, nil)
}

func (s *service) DeleteSocialProgramSubscription(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid subscription ID format"}, nil)
	}

	_, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Subscription not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch subscription", nil, nil)
	}

	if err := s.repo.DeleteSocialProgramSubscription(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to delete subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete subscription", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Subscription deleted successfully", nil, nil)
}
