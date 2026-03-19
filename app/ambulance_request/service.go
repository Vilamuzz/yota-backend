package ambulance_request

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceRequest(ctx context.Context, queryParams AmbulanceRequestQueryParams) pkg.Response
	GetAmbulanceRequestByID(ctx context.Context, id string) pkg.Response
	CreateAmbulanceRequest(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response
	UpdateAmbulanceRequest(ctx context.Context, id string, payload UpdateAmbulanceRequest) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) ListAmbulanceRequest(ctx context.Context, queryParams AmbulanceRequestQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}
	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	ambulanceRequests, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to list ambulance requests", nil, nil)
	}
	hasNext := len(ambulanceRequests) > queryParams.Limit
	if hasNext {
		ambulanceRequests = ambulanceRequests[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulanceRequests) > 0 {
		lastRequest := ambulanceRequests[len(ambulanceRequests)-1]
		nextCursor = pkg.EncodeCursor(lastRequest.CreatedAt, lastRequest.ID)
	}
	if hasPrev && len(ambulanceRequests) > 0 {
		firstRequest := ambulanceRequests[0]
		prevCursor = pkg.EncodeCursor(firstRequest.CreatedAt, firstRequest.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toAmbulanceRequestsToListResponse(ambulanceRequests, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) GetAmbulanceRequestByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulanceRequest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulance request not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance request", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Success", nil, ambulanceRequest.toAmbulanceRequestResponse())
}

func (s *service) CreateAmbulanceRequest(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.ApplicantName == "" {
		errValidation["applicant_name"] = "Applicant name is required"
	}
	if payload.ApplicantPhone == "" {
		errValidation["applicant_phone"] = "Applicant phone is required"
	}
	if payload.ApplicantAddress == "" {
		errValidation["applicant_address"] = "Applicant address is required"
	}
	if payload.Date == "" {
		errValidation["date"] = "Date is required"
	}
	if payload.Reason == "" {
		errValidation["reason"] = "Reason is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	request := AmbulanceRequest{
		ID:               uuid.New().String(),
		UserID:           payload.UserID,
		ApplicantName:    payload.ApplicantName,
		ApplicantPhone:   payload.ApplicantPhone,
		ApplicantAddress: payload.ApplicantAddress,
		Date:             time.Now(),
		Reason:           payload.Reason,
		Status:           StatusPending,
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance request", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance request created successfully", nil, nil)
}

func (s *service) UpdateAmbulanceRequest(ctx context.Context, id string, payload UpdateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid ambulance request ID format"}, nil)
	}

	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulance request not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance request", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})
	if payload.Status != "" && payload.Status != string(StatusPending) && payload.Status != string(StatusApproved) && payload.Status != string(StatusRejected) {
		errValidation["status"] = "Invalid status value"
	} else if payload.Status != "" {
		updateData["status"] = payload.Status
	}
	if payload.Status == string(StatusRejected) && payload.RejectReason == "" {
		errValidation["reject_reason"] = "Reject reason is required when status is rejected"
	} else if payload.Status == string(StatusRejected) {
		updateData["reject_reason"] = payload.RejectReason
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}
	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "No data to update", nil, nil)
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update ambulance request", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance request updated successfully", nil, nil)
}
