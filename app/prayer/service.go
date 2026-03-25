package prayer

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
	FindPrayerByID(ctx context.Context, id string, userID string) pkg.Response
	ListPrayers(ctx context.Context, params PrayerQueryParams, userID string) pkg.Response
	ListReportedPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response
	DeletePrayer(ctx context.Context, id string) pkg.Response
	PrayerAmen(ctx context.Context, payload PrayerAmenRequest, userID string) pkg.Response
	CreateReportPrayer(ctx context.Context, payload ReportPrayerRequest, userID string) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) PrayerAmen(ctx context.Context, payload PrayerAmenRequest, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.repo.FindByID(ctx, payload.PrayerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Prayer not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}

	rowsAffected, err := s.repo.DeleteAmen(ctx, payload.PrayerID, userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "prayer.service",
			"prayer_id": payload.PrayerID,
			"user_id":   userID,
		}).WithError(err).Error("failed to delete amen")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete amen", nil, nil)
	}

	// If deletion was successful, return success
	if rowsAffected > 0 {
		return pkg.NewResponse(http.StatusOK, "Amen deleted successfully", nil, map[string]interface{}{
			"is_amen": false,
		})
	}

	// If no amen was deleted, create a new one
	amen := &PrayerAmen{
		ID:       uuid.New().String(),
		PrayerID: payload.PrayerID,
		UserID:   userID,
	}
	if err := s.repo.CreateAmen(ctx, amen); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "prayer.service",
			"prayer_id": payload.PrayerID,
			"user_id":   userID,
		}).WithError(err).Error("failed to create amen")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create amen", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Amen created successfully", nil, map[string]interface{}{
		"is_amen": true,
	})
}

func (s *service) CreateReportPrayer(ctx context.Context, payload ReportPrayerRequest, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	_, err := s.repo.FindByID(ctx, payload.PrayerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Prayer not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}

	if payload.Reason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Validation Error", map[string]string{
			"reason": "Reason is required",
		}, nil)
	}

	_, err = s.repo.FindReport(ctx, map[string]interface{}{
		"prayer_id": payload.PrayerID,
		"user_id":   userID,
	})
	if err == nil {
		return pkg.NewResponse(http.StatusOK, "Prayer reported successfully", nil, nil)
	}

	report := &PrayerReport{
		ID:       uuid.New().String(),
		PrayerID: payload.PrayerID,
		UserID:   userID,
		Reason:   payload.Reason,
	}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "prayer.service",
			"prayer_id": payload.PrayerID,
			"user_id":   userID,
		}).WithError(err).Error("failed to create prayer report")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create report", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer reported successfully", nil, nil)
}

func (s *service) FindPrayerByID(ctx context.Context, id string, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Prayer not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}

	if userID != "" {
		isAmen, err := s.repo.ExistsAmen(ctx, prayer.ID, userID)
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to determine amen status", nil, nil)
		}
		prayer.IsAmen = isAmen
	}

	return pkg.NewResponse(http.StatusOK, "Prayer found successfully", nil, prayer.toPrayerResponse())
}

func (s *service) ListPrayers(ctx context.Context, params PrayerQueryParams, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	prayers, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayers", nil, nil)
	}

	hasMore := len(prayers) > params.Limit
	if hasMore {
		prayers = prayers[:params.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(prayers)-1; i < j; i, j = i+1, j-1 {
			prayers[i], prayers[j] = prayers[j], prayers[i]
		}
	}

	if userID != "" && len(prayers) > 0 {
		prayerIDs := make([]string, 0, len(prayers))
		for _, prayer := range prayers {
			prayerIDs = append(prayerIDs, prayer.ID)
		}

		amenPrayerIDs, err := s.repo.FindAmenPrayerIDs(ctx, userID, prayerIDs)
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to determine amen status", nil, nil)
		}

		for i := range prayers {
			prayers[i].IsAmen = amenPrayerIDs[prayers[i].ID]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(prayers) > 0 {
		first := prayers[0]
		last := prayers[len(prayers)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Prayers found successfully", nil, toPrayerListResponse(prayers, pagination))
}

func (s *service) ListReportedPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit":    params.Limit,
		"reported": true,
	}
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	prayers, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayers", nil, nil)
	}

	hasMore := len(prayers) > params.Limit
	if hasMore {
		prayers = prayers[:params.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(prayers)-1; i < j; i, j = i+1, j-1 {
			prayers[i], prayers[j] = prayers[j], prayers[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(prayers) > 0 {
		first := prayers[0]
		last := prayers[len(prayers)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Prayers found successfully", nil, toPrayerReportedListResponse(prayers, pagination))
}

func (s *service) DeletePrayer(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Prayer not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer deleted successfully", nil, nil)
}
