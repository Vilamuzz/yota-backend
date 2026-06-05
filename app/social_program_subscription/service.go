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
	GetSocialProgramSubscriptionList(ctx context.Context, socialProgramID string, params SocialProgramSubscriptionQueryParams) pkg.Response
	GetSocialProgramSubscriptionByID(ctx context.Context, id string, isSubscriber bool) pkg.Response
	CreateSocialProgramSubscription(ctx context.Context, accountID string, socialProgramID string) pkg.Response
	UpdateSocialProgramSubscription(ctx context.Context, id string, req UpdateSocialProgramSubscriptionRequest) pkg.Response
	DeactivateSocialProgramSubscription(ctx context.Context, id string, accountID string) pkg.Response
	GetSubscribers(ctx context.Context, params SocialProgramSubscriptionQueryParams) pkg.Response
	GetSubscriberByID(ctx context.Context, id string) pkg.Response
	GetSocialProgramSubscriptionsByAccountID(ctx context.Context, accountID string, params SocialProgramSubscriptionQueryParams) pkg.Response
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

func (s *service) GetSocialProgramSubscriptionList(ctx context.Context, socialProgramID string, params SocialProgramSubscriptionQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"social_program_id": "Format ID program sosial tidak valid"}, nil)
	}

	if _, err := s.socialProgramRepo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID}); err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	options := map[string]interface{}{
		"social_program_id": socialProgramID,
	}

	return s.getSubscriptionList(ctx, options, params)
}

func (s *service) GetSocialProgramSubscriptionsByAccountID(ctx context.Context, accountID string, params SocialProgramSubscriptionQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	options := map[string]interface{}{
		"account_id": accountID,
		"limit":      params.Limit,
		"page":       params.Page,
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.Search != "" {
		options["search"] = params.Search
	}

	total, err := s.repo.CountSocialProgramSubscriptions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
			"options":   options,
		}).WithError(err).Error("failed to count subscriptions")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghitung data langganan", nil, nil)
	}

	subscriptions, err := s.repo.FindAllSocialProgramSubscriptions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
			"options":   options,
		}).WithError(err).Error("failed to fetch subscriptions")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data langganan", nil, nil)
	}

	var subscriptionIDs []string
	for _, sub := range subscriptions {
		subscriptionIDs = append(subscriptionIDs, sub.ID.String())
	}

	donationsMap, err := s.repo.GetTotalDonationBySubscriptionIDs(ctx, subscriptionIDs)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to fetch donations map")
		if donationsMap == nil {
			donationsMap = make(map[string]float64)
		}
	}

	pagination := pkg.OffsetPagination{
		Page:  params.Page,
		Limit: params.Limit,
		Total: int64(total),
	}

	return pkg.NewResponse(http.StatusOK, "Data langganan berhasil ditemukan", nil, toSubscriberSubscriptionListResponse(subscriptions, pagination, donationsMap))
}

func (s *service) getSubscriptionList(ctx context.Context, options map[string]interface{}, params SocialProgramSubscriptionQueryParams) pkg.Response {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	options["limit"] = params.Limit
	options["page"] = params.Page
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.Search != "" {
		options["search"] = params.Search
	}

	total, err := s.repo.CountSocialProgramSubscriptions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
			"options":   options,
		}).WithError(err).Error("failed to count subscriptions")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghitung data langganan", nil, nil)
	}

	subscriptions, err := s.repo.FindAllSocialProgramSubscriptions(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
			"options":   options,
		}).WithError(err).Error("failed to fetch subscriptions")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data langganan", nil, nil)
	}

	pagination := pkg.OffsetPagination{
		Page:  params.Page,
		Limit: params.Limit,
		Total: int64(total),
	}

	return pkg.NewResponse(http.StatusOK, "Data langganan berhasil ditemukan", nil, toSocialProgramSubscriptionListResponse(subscriptions, pagination))
}

func (s *service) GetSocialProgramSubscriptionByID(ctx context.Context, id string, isSubscriber bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID langganan tidak valid"}, nil)
	}

	subscription, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Langganan tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to fetch subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data langganan", nil, nil)
	}

	if isSubscriber {
		return pkg.NewResponse(http.StatusOK, "Data langganan berhasil ditemukan", nil, subscription.toSubscriberSubscriptionResponse(subscription.TotalDonation))
	}
	return pkg.NewResponse(http.StatusOK, "Data langganan berhasil ditemukan", nil, subscription.toSocialProgramSubscriptionResponse())
}

func (s *service) CreateSocialProgramSubscription(ctx context.Context, accountID string, socialProgramID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"social_program_id": "Format ID program sosial tidak valid"}, nil)
	}

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}

	_, err := s.socialProgramRepo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	existing, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{
		"social_program_id": socialProgramID,
		"account_id":        accountID,
	})
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.WithFields(logrus.Fields{
			"component":         "social_program_subscription.service",
			"social_program_id": socialProgramID,
			"account_id":        accountID,
		}).WithError(err).Error("failed to check for existing subscription")
	}

	if existing != nil {
		if existing.Status == StatusActive {
			return pkg.NewResponse(http.StatusConflict, "Langganan aktif sudah ada untuk program sosial ini", nil, nil)
		}

		updates := map[string]interface{}{
			"status":     StatusActive,
			"updated_at": time.Now(),
		}
		if err := s.repo.UpdateSocialProgramSubscription(ctx, existing.ID.String(), updates); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":       "social_program_subscription.service",
				"subscription_id": existing.ID,
			}).WithError(err).Error("failed to reactivate subscription")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengaktifkan kembali langganan", nil, nil)
		}

		existing.Status = StatusActive
		return pkg.NewResponse(http.StatusOK, "Langganan berhasil diaktifkan kembali", nil, existing.toSocialProgramSubscriptionResponse())
	}

	now := time.Now()
	subscription := &SocialProgramSubscription{
		ID:              uuid.New(),
		SocialProgramID: uuid.MustParse(socialProgramID),
		AccountID:       uuid.MustParse(accountID),
		Status:          StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateSocialProgramSubscription(ctx, subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to create subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat langganan", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Langganan berhasil dibuat", nil, subscription.toSocialProgramSubscriptionResponse())
}

func (s *service) UpdateSocialProgramSubscription(ctx context.Context, id string, req UpdateSocialProgramSubscriptionRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID langganan tidak valid"}, nil)
	}

	_, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Langganan tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data langganan", nil, nil)
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
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui langganan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Langganan berhasil diperbarui", nil, nil)
}

func (s *service) DeactivateSocialProgramSubscription(ctx context.Context, id string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID langganan tidak valid"}, nil)
	}

	subscription, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Langganan tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to fetch subscription for deactivation")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data langganan", nil, nil)
	}

	if accountID != "" && subscription.AccountID.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Akses ditolak", nil, nil)
	}

	updates := map[string]interface{}{
		"status":     StatusInactive,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateSocialProgramSubscription(ctx, id, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":       "social_program_subscription.service",
			"subscription_id": id,
		}).WithError(err).Error("failed to deactivate subscription")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menonaktifkan langganan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Langganan berhasil dinonaktifkan", nil, nil)
}

func (s *service) GetSubscribers(ctx context.Context, params SocialProgramSubscriptionQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	options := map[string]interface{}{
		"limit": params.Limit,
		"page":  params.Page,
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}
	if params.Search != "" {
		options["search"] = params.Search
	}

	total, err := s.repo.CountSubscribers(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to count subscribers")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghitung data pelanggan", nil, nil)
	}

	subscriptions, err := s.repo.FindAllSubscribers(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to fetch subscribers")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pelanggan", nil, nil)
	}

	var accountIDs []string
	for _, sub := range subscriptions {
		accountIDs = append(accountIDs, sub.AccountID.String())
	}

	statsMap, err := s.repo.GetSubscriberStats(ctx, accountIDs)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program_subscription.service",
		}).WithError(err).Error("failed to fetch subscriber stats")
		if statsMap == nil {
			statsMap = make(map[string]SubscriberStats)
		}
	}

	pagination := pkg.OffsetPagination{
		Page:  params.Page,
		Limit: params.Limit,
		Total: int64(total),
	}

	return pkg.NewResponse(http.StatusOK, "Data pelanggan berhasil ditemukan", nil, toSubscriptionsListResponse(subscriptions, pagination, statsMap))
}

func (s *service) GetSubscriberByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID pelanggan tidak valid"}, nil)
	}

	subscription, err := s.repo.FindOneSocialProgramSubscription(ctx, map[string]interface{}{"account_id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Pelanggan tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_subscription.service",
			"account_id": id,
		}).WithError(err).Error("failed to fetch subscriber")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data pelanggan", nil, nil)
	}

	statsMap, err := s.repo.GetSubscriberStats(ctx, []string{id})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "social_program_subscription.service",
			"account_id": id,
		}).WithError(err).Error("failed to fetch subscriber stats")
		if statsMap == nil {
			statsMap = make(map[string]SubscriberStats)
		}
	}

	stats := statsMap[id]
	return pkg.NewResponse(http.StatusOK, "Data pelanggan berhasil ditemukan", nil, subscription.toSubscribersResponse(stats))
}
