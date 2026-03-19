package ambulance_history

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response
	CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response
	UpdateAmbulanceHistory(ctx context.Context, id int, payload UpdateAmbulanceHistoryRequest) pkg.Response
	DeleteAmbulanceHistory(ctx context.Context, id int) pkg.Response
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
	if queryParams.AmbulanceID != 0 {
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
		nextCursor = pkg.EncodeCursor(lastHistory.CreatedAt, strconv.Itoa(lastHistory.ID))
	}
	if hasPrev && len(histories) > 0 {
		firstHistory := histories[0]
		prevCursor = pkg.EncodeCursor(firstHistory.CreatedAt, strconv.Itoa(firstHistory.ID))
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toAmbulanceHistoriesToListResponse(histories, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.AmbulanceID == 0 {
		errValidation["ambulance_id"] = "Ambulance ID is required"
	} else if payload.AmbulanceID != 0 {
		_, err := s.ambulanceRepo.FindByID(ctx, payload.AmbulanceID)
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
	} else if payload.ServiceCategory != SocialService && payload.ServiceCategory != EmergencyService && payload.ServiceCategory != OtherService {
		errValidation["service_category"] = "Invalid service category"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	now := time.Now()
	ambulanceHistory := AmbulanceHistory{
		AmbulanceID:     payload.AmbulanceID,
		ServiceCategory: payload.ServiceCategory,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, ambulanceHistory); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance history", nil, nil)
	}
	return pkg.NewResponse(http.StatusCreated, "Ambulance history created successfully", nil, nil)
}

func (s *service) UpdateAmbulanceHistory(ctx context.Context, id int, payload UpdateAmbulanceHistoryRequest) pkg.Response {
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

func (s *service) DeleteAmbulanceHistory(ctx context.Context, id int) pkg.Response {
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
