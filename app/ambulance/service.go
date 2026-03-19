package ambulance

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type Service interface {
	CreateAmbulance(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response
	FindAmbulanceById(ctx context.Context, id int) pkg.Response
	ListAmbulance(ctx context.Context, queryParams AmbulanceQueryParams) pkg.Response
	UpdateAmbulance(ctx context.Context, id int, payload UpdateAmbulanceRequest) pkg.Response
	DeleteAmbulance(ctx context.Context, id int) pkg.Response
}

type service struct {
	repo     Repository
	timeout  time.Duration
	s3Client s3_pkg.Client
}

func NewService(repo Repository, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{repo: repo, s3Client: s3Client, timeout: timeout}
}

func (s *service) CreateAmbulance(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.Image == nil {
		errValidation["image"] = "Image is required"
	}
	if payload.PlateNumber == "" {
		errValidation["plate_number"] = "Plate number is required"
	}
	if payload.Phone == "" {
		errValidation["phone"] = "Phone is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var imageURL string
	if payload.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.Image, "ambulances")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		imageURL = uploadedURL
	}

	now := time.Now()
	ambulance := Ambulance{
		PlateNumber: payload.PlateNumber,
		Phone:       payload.Phone,
		ImageURL:    imageURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ambulance); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance created successfully", nil, nil)
}

func (s *service) FindAmbulanceById(ctx context.Context, id int) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulance, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find ambulance", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance found successfully", nil, ambulance)
}

func (s *service) ListAmbulance(ctx context.Context, queryParams AmbulanceQueryParams) pkg.Response {
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

	ambulances, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find ambulances", nil, nil)
	}
	hasNext := len(ambulances) > queryParams.Limit
	if hasNext {
		ambulances = ambulances[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulances) > 0 {
		lastAmbulance := ambulances[len(ambulances)-1]
		nextCursor = pkg.EncodeCursor(lastAmbulance.CreatedAt, strconv.Itoa(lastAmbulance.ID))
	}
	if hasPrev && len(ambulances) > 0 {
		firstAmbulance := ambulances[0]
		prevCursor = pkg.EncodeCursor(firstAmbulance.CreatedAt, strconv.Itoa(firstAmbulance.ID))
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toAmbulanceListResponse(ambulances, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) UpdateAmbulance(ctx context.Context, id int, payload UpdateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulance, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find ambulance", nil, nil)
	}

	updateData := make(map[string]interface{})
	if payload.PlateNumber != "" {
		updateData["plate_number"] = payload.PlateNumber
	}
	if payload.Phone != "" {
		updateData["phone"] = payload.Phone
	}
	if payload.Image != nil {
		objectName := s3_pkg.ExtractObjectNameFromURL(ambulance.ImageURL)
		if objectName != "" {
			_ = s.s3Client.DeleteFile(ctx, objectName)
		}

		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.Image, "ambulances")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		updateData["image_url"] = uploadedURL
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update ambulance", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance updated successfully", nil, nil)
}

func (s *service) DeleteAmbulance(ctx context.Context, id int) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete ambulance", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulance deleted successfully", nil, nil)
}
