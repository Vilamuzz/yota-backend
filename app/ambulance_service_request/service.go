package ambulance_service_request

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	GetAmbulanceServiceRequestByID(ctx context.Context, id string) pkg.Response
	CreateAmbulanceServiceRequest(ctx context.Context, payload CreateAmbulanceServiceRequest) pkg.Response
	UpdateAmbulanceServiceRequest(ctx context.Context, id string, payload UpdateAmbulanceServiceRequest) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestQueryParams) pkg.Response {
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

	ambulanceServiceRequests, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to list ambulance requests", nil, nil)
	}
	hasNext := len(ambulanceServiceRequests) > queryParams.Limit
	if hasNext {
		ambulanceServiceRequests = ambulanceServiceRequests[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulanceServiceRequests) > 0 {
		lastRequest := ambulanceServiceRequests[len(ambulanceServiceRequests)-1]
		nextCursor = pkg.EncodeCursor(lastRequest.CreatedAt, lastRequest.ID.String())
	}
	if hasPrev && len(ambulanceServiceRequests) > 0 {
		firstRequest := ambulanceServiceRequests[0]
		prevCursor = pkg.EncodeCursor(firstRequest.CreatedAt, firstRequest.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toAmbulanceServiceRequestsToListResponse(ambulanceServiceRequests, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) GetAmbulanceServiceRequestByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulanceServiceRequest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulance request not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to get ambulance request", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Success", nil, ambulanceServiceRequest.toAmbulanceServiceRequestResponse())
}

func (s *service) CreateAmbulanceServiceRequest(ctx context.Context, payload CreateAmbulanceServiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.ApplicantName == "" {
		errValidation["applicantName"] = "Applicant name is required"
	}
	if payload.ApplicantPhone == "" {
		errValidation["applicantPhone"] = "Applicant phone is required"
	}
	if payload.ApplicantAddress == "" {
		errValidation["applicantAddress"] = "Applicant address is required"
	}
	if payload.RequestDate == "" {
		errValidation["requestDate"] = "Request date is required"
	}
	if payload.RequestReason == "" {
		errValidation["requestReason"] = "Request reason is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	request := AmbulanceServiceRequest{
		ID:               uuid.New(),
		AccountID:        uuid.MustParse(payload.AccountID),
		ApplicantName:    payload.ApplicantName,
		ApplicantPhone:   payload.ApplicantPhone,
		ApplicantAddress: payload.ApplicantAddress,
		RequestDate:      time.Now(),
		RequestReason:    payload.RequestReason,
		Status:           StatusPending,
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance request", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance request created successfully", nil, nil)
}

func (s *service) UpdateAmbulanceServiceRequest(ctx context.Context, id string, payload UpdateAmbulanceServiceRequest) pkg.Response {
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
	if payload.Status == string(StatusRejected) && payload.RejectionReason == "" {
		errValidation["rejectionReason"] = "Rejection reason is required when status is rejected"
	} else if payload.Status == string(StatusRejected) {
		updateData["rejection_reason"] = payload.RejectionReason
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
