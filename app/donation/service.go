package donation

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
)

type Service interface {
	ListPublished(ctx context.Context, queryParams DonationQueryParams) pkg.Response
	GetPublishedByID(ctx context.Context, id string) pkg.Response
	List(ctx context.Context, queryParams DonationQueryParams) pkg.Response
	GetByID(ctx context.Context, id string) pkg.Response
	CreateDonation(ctx context.Context, donation DonationRequest) pkg.Response
	UpdateDonation(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response
	DeleteDonation(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo     Repository
	s3Client s3_pkg.Client
	timeout  time.Duration
}

func NewService(repo Repository, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:     repo,
		s3Client: s3Client,
		timeout:  timeout,
	}
}

func (s *service) ListPublished(ctx context.Context, queryParams DonationQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Set default limit
	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.Category != "" {
		options["category"] = queryParams.Category
	}
	if queryParams.Cursor != "" {
		options["cursor"] = queryParams.Cursor
	}

	donations, err := s.repo.FindPublished(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch donations", nil, nil)
	}

	// Check if there are more results
	hasNext := len(donations) > queryParams.Limit
	if hasNext {
		// Remove the extra item
		donations = donations[:queryParams.Limit]
	}

	// Generate cursors
	var nextCursor, prevCursor string
	hasPrev := queryParams.Cursor != ""

	if hasNext && len(donations) > 0 {
		lastDonation := donations[len(donations)-1]
		nextCursor = pkg.EncodeCursor(lastDonation.CreatedAt, lastDonation.ID)
	}

	if hasPrev && len(donations) > 0 {
		firstDonation := donations[0]
		prevCursor = pkg.EncodeCursor(firstDonation.CreatedAt, firstDonation.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"donations": donations,
		"pagination": map[string]interface{}{
			"next_cursor": nextCursor,
			"prev_cursor": prevCursor,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"limit":       queryParams.Limit,
		},
	})
}

func (s *service) List(ctx context.Context, queryParams DonationQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Set default limit
	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.Category != "" {
		options["category"] = queryParams.Category
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.Cursor != "" {
		options["cursor"] = queryParams.Cursor
	}

	donations, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch donations", nil, nil)
	}

	// Check if there are more results
	hasNext := len(donations) > queryParams.Limit
	if hasNext {
		// Remove the extra item
		donations = donations[:queryParams.Limit]
	}

	// Generate cursors
	var nextCursor, prevCursor string
	hasPrev := queryParams.Cursor != ""

	if hasNext && len(donations) > 0 {
		lastDonation := donations[len(donations)-1]
		nextCursor = pkg.EncodeCursor(lastDonation.CreatedAt, lastDonation.ID)
	}

	if hasPrev && len(donations) > 0 {
		firstDonation := donations[0]
		prevCursor = pkg.EncodeCursor(firstDonation.CreatedAt, firstDonation.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"donations": donations,
		"pagination": map[string]interface{}{
			"next_cursor": nextCursor,
			"prev_cursor": prevCursor,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"limit":       queryParams.Limit,
		},
	})
}

func (s *service) GetPublishedByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindPublishedByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation)
}

func (s *service) GetByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation)
}

func (s *service) CreateDonation(ctx context.Context, req DonationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validation
	errValidation := make(map[string]string)

	// Validate title
	if req.Title == "" {
		errValidation["title"] = "Title is required"
	} else if len(req.Title) < 3 {
		errValidation["title"] = "Title must be at least 3 characters"
	} else if len(req.Title) > 200 {
		errValidation["title"] = "Title must not exceed 200 characters"
	}

	// Validate description
	if req.Description == "" {
		errValidation["description"] = "Description is required"
	} else if len(req.Description) < 10 {
		errValidation["description"] = "Description must be at least 10 characters"
	} else if len(req.Description) > 2000 {
		errValidation["description"] = "Description must not exceed 2000 characters"
	}

	// Validate category
	if req.Category == "" {
		errValidation["category"] = "Category is required"
	} else if req.Category != CategoryEducation && req.Category != CategoryHealth && req.Category != CategoryEnvironment {
		errValidation["category"] = "Invalid category. Must be: education, health, or environment"
	}

	// Validate fund target
	if req.FundTarget <= 0 {
		errValidation["fund_target"] = "Fund target must be greater than 0"
	}

	// Validate end date
	if req.DateEnd.IsZero() {
		errValidation["date_end"] = "End date is required"
	} else if req.DateEnd.Before(time.Now()) {
		errValidation["date_end"] = "End date must be in the future"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Handle image upload
	var imageURL string
	if req.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.Image, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		imageURL = uploadedURL
	}

	// Handle status
	status := StatusDraft
	if req.Status {
		status = StatusActive
	}

	// Create donation
	timeNow := time.Now()
	donation := &Donation{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    imageURL,
		Category:    req.Category,
		FundTarget:  req.FundTarget,
		Status:      status,
		DateEnd:     req.DateEnd,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	if err := s.repo.Create(ctx, donation); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Donation successfully created", nil, donation)
}

func (s *service) UpdateDonation(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	// Check if donation exists
	existingDonation, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	// Validation
	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	// Validate and set title
	if req.Title != "" {
		if len(req.Title) < 3 {
			errValidation["title"] = "Title must be at least 3 characters"
		} else if len(req.Title) > 200 {
			errValidation["title"] = "Title must not exceed 200 characters"
		} else {
			updateData["title"] = req.Title
		}
	}

	// Validate and set description
	if req.Description != "" {
		if len(req.Description) < 10 {
			errValidation["description"] = "Description must be at least 10 characters"
		} else if len(req.Description) > 2000 {
			errValidation["description"] = "Description must not exceed 2000 characters"
		} else {
			updateData["description"] = req.Description
		}
	}

	// Validate and set category
	if req.Category != "" {
		if req.Category != CategoryEducation && req.Category != CategoryHealth && req.Category != CategoryEnvironment {
			errValidation["category"] = "Invalid category. Must be: education, health, or environment"
		} else {
			updateData["category"] = req.Category
		}
	}

	// Validate and set fund target
	if req.FundTarget > 0 {
		updateData["fund_target"] = req.FundTarget
	} else if req.FundTarget < 0 {
		errValidation["fund_target"] = "Fund target must be greater than 0"
	}

	// Validate and set status
	if req.Status != nil {
		status := StatusDraft
		if *req.Status {
			status = StatusActive
		}

		if status == StatusActive && time.Now().After(existingDonation.DateEnd) {
			errValidation["status"] = "Cannot activate donation that has already ended"
		} else {
			updateData["status"] = status
		}
	}

	// Validate and set end date
	if !req.DateEnd.IsZero() {
		if req.DateEnd.Before(time.Now()) {
			errValidation["date_end"] = "End date must be in the future"
		} else {
			updateData["date_end"] = req.DateEnd
		}
	}

	// Validate and set image
	if req.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.Image, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		updateData["image_url"] = uploadedURL
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	// Set updated_at
	updateData["updated_at"] = time.Now()

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Donation updated successfully", nil, nil)
}

func (s *service) DeleteDonation(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	// Check if donation exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Donation deleted successfully", nil, nil)
}
