package donation

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type Service interface {
	FetchAllDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response
	FetchDonationByID(ctx context.Context, id string) pkg.Response
	Create(ctx context.Context, donation DonationRequest) pkg.Response
	Update(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response
	Delete(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *service) FetchAllDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response {
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

	donations, err := s.repo.FetchAllDonations(ctx, options)
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

func (s *service) FetchDonationByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FetchDonationByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation)
}

func (s *service) Create(ctx context.Context, req DonationRequest) pkg.Response {
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

	// Create donation
	timeNow := time.Now()
	donation := &Donation{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		Category:    req.Category,
		FundTarget:  req.FundTarget,
		Status:      StatusActive,
		DateEnd:     req.DateEnd,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	if err := s.repo.Create(ctx, donation); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Donation successfully created", nil, donation)
}

func (s *service) Update(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	// Check if donation exists
	existingDonation, err := s.repo.FetchDonationByID(ctx, id)
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
	if req.Status != "" {
		if req.Status != StatusActive && req.Status != StatusInactive && req.Status != StatusCompleted {
			errValidation["status"] = "Invalid status. Must be: active, inactive, or completed"
		} else {
			// Prevent changing status to active if date has passed
			if req.Status == StatusActive && time.Now().After(existingDonation.DateEnd) {
				errValidation["status"] = "Cannot activate donation that has already ended"
			} else {
				updateData["status"] = req.Status
			}
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
	if req.Image != "" {
		updateData["image"] = req.Image
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

func (s *service) Delete(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	// Check if donation exists
	_, err := s.repo.FetchDonationByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Donation deleted successfully", nil, nil)
}
