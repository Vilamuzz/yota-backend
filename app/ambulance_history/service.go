package ambulance_history

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response
	AmbulanceHistorySummary(ctx context.Context, ambulanceID string, params AmbulanceSummaryQueryParams) pkg.Response
	CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response
	UpdateAmbulanceHistory(ctx context.Context, id string, payload UpdateAmbulanceHistoryRequest) pkg.Response
	DeleteAmbulanceHistory(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo          Repository
	ambulanceRepo ambulance.Repository
	timeout       time.Duration
}

func NewService(repo Repository, ambulanceRepo ambulance.Repository, timeout time.Duration) Service {
	return &service{
		repo:          repo,
		ambulanceRepo: ambulanceRepo,
		timeout:       timeout,
	}
}

func (s *service) ListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.AmbulanceID != "" {
		options["ambulance_id"] = queryParams.AmbulanceID
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	histories, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(500, "Failed to list ambulance history", nil, nil)
	}

	hasNext := len(histories) > queryParams.Limit
	if hasNext {
		histories = histories[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(histories) > 0 {
		lastHistory := histories[len(histories)-1]
		nextCursor = pkg.EncodeCursor(lastHistory.CreatedAt, lastHistory.ID.String())
	}
	if hasPrev && len(histories) > 0 {
		firstHistory := histories[0]
		prevCursor = pkg.EncodeCursor(firstHistory.CreatedAt, firstHistory.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toAmbulanceHistoriesToListResponse(histories, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) AmbulanceHistorySummary(ctx context.Context, ambulanceID string, params AmbulanceSummaryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if ambulanceID == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Ambulance ID is required", nil, nil)
	}

	var startDate, endDate *time.Time
	now := time.Now()

	// Default to all_time when period is not specified
	if params.Period == "" {
		params.Period = PeriodAllTime
	}

	switch params.Period {
	case PeriodThisWeek:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday → treat as end of week
		}
		start := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
		end := start.AddDate(0, 0, 6).Add(24*time.Hour - time.Nanosecond)
		startDate, endDate = &start, &end

	case PeriodThisMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		startDate, endDate = &start, &end

	case PeriodThisYear:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		startDate, endDate = &start, &end

	case PeriodCustom:
		if params.StartDate == "" || params.EndDate == "" {
			return pkg.NewResponse(http.StatusBadRequest,
				"startDate and endDate are required when period is \"custom\"", nil, nil)
		}
		parsedStart, err := time.ParseInLocation("2006-01-02", params.StartDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("invalid startDate format: %s (expected YYYY-MM-DD)", params.StartDate), nil, nil)
		}
		parsedEnd, err := time.ParseInLocation("2006-01-02", params.EndDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("invalid endDate format: %s (expected YYYY-MM-DD)", params.EndDate), nil, nil)
		}
		parsedEnd = parsedEnd.Add(24*time.Hour - time.Nanosecond) // inclusive end
		if parsedStart.After(parsedEnd) {
			return pkg.NewResponse(http.StatusBadRequest, "startDate must be before endDate", nil, nil)
		}
		startDate, endDate = &parsedStart, &parsedEnd

	case PeriodAllTime:
		// no date filter

	default:
		return pkg.NewResponse(http.StatusBadRequest,
			"invalid period; accepted values: all_time, this_week, this_month, this_year, custom", nil, nil)
	}

	counts, err := s.repo.GetSummary(ctx, ambulanceID, startDate, endDate)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance history summary", nil, nil)
	}

	var total int64
	for _, c := range counts {
		total += c.Count
	}

	summary := SummaryResponse{
		Total:      total,
		Categories: counts,
		Period:     string(params.Period),
	}
	if startDate != nil {
		summary.StartDate = startDate.Format("2006-01-02")
	}
	if endDate != nil {
		summary.EndDate = endDate.Format("2006-01-02")
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, summary)
}

func (s *service) CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.AmbulanceID == "" {
		errValidation["ambulance_id"] = "Ambulance ID is required"
	} else if payload.AmbulanceID != "" {
		_, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": payload.AmbulanceID})
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				errValidation["ambulance_id"] = "Ambulance not found"
			} else {
				return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance", nil, nil)
			}
		}
	}
	if payload.ServiceCategory == "" {
		errValidation["service_category"] = "Service category is required"
	} else if payload.ServiceCategory != SocialService && payload.ServiceCategory != EmergencyService && payload.ServiceCategory != OtherService && payload.ServiceCategory != MortuaryService && payload.ServiceCategory != PatientService {
		errValidation["service_category"] = "Invalid service category"
	}

	if payload.DriverID == "" {
		errValidation["driver_id"] = "Driver ID is required"
	} else if _, err := uuid.Parse(payload.DriverID); err != nil {
		errValidation["driver_id"] = "Invalid Driver ID format"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	now := time.Now()
	ambulanceHistory := AmbulanceHistory{
		AmbulanceID:     uuid.MustParse(payload.AmbulanceID),
		DriverID:        uuid.MustParse(payload.DriverID),
		ServiceCategory: payload.ServiceCategory,
		Note:            payload.Note,
		CreatedAt:       now,
	}
	if err := s.repo.Create(ctx, ambulanceHistory); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance history", nil, nil)
	}
	return pkg.NewResponse(http.StatusCreated, "Ambulance history created successfully", nil, nil)
}

func (s *service) UpdateAmbulanceHistory(ctx context.Context, id string, payload UpdateAmbulanceHistoryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulanceHistory, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Ambulance history not found", nil, nil)
	}

	if err := s.repo.Update(ctx, ambulanceHistory); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update ambulance history", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance history updated successfully", nil, nil)
}

func (s *service) DeleteAmbulanceHistory(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulance history not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance history", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete ambulance history", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance history deleted successfully", nil, nil)
}
