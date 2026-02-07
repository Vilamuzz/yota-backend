package gallery

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type Service interface {
	FetchAllGalleries(ctx context.Context, queryParams GalleryQueryParams) pkg.Response
	FetchGalleryByID(ctx context.Context, id string, incrementView bool) pkg.Response
	Create(ctx context.Context, req GalleryRequest) pkg.Response
	Update(ctx context.Context, id string, req UpdateGalleryRequest) pkg.Response
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

func (s *service) FetchAllGalleries(ctx context.Context, queryParams GalleryQueryParams) pkg.Response {
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

	galleries, err := s.repo.FetchAllGalleries(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch galleries", nil, nil)
	}

	// Check if there are more results
	hasNext := len(galleries) > queryParams.Limit
	if hasNext {
		galleries = galleries[:queryParams.Limit]
	}

	// Generate cursors
	var nextCursor, prevCursor string
	hasPrev := queryParams.Cursor != ""

	if hasNext && len(galleries) > 0 {
		lastGallery := galleries[len(galleries)-1]
		nextCursor = pkg.EncodeCursor(lastGallery.CreatedAt, lastGallery.ID)
	}

	if hasPrev && len(galleries) > 0 {
		firstGallery := galleries[0]
		prevCursor = pkg.EncodeCursor(firstGallery.CreatedAt, firstGallery.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"galleries": galleries,
		"pagination": map[string]interface{}{
			"next_cursor": nextCursor,
			"prev_cursor": prevCursor,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"limit":       queryParams.Limit,
		},
	})
}

func (s *service) FetchGalleryByID(ctx context.Context, id string, incrementView bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}

	gallery, err := s.repo.FetchGalleryByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
	}

	// Increment views if requested (for public access)
	if incrementView {
		go s.repo.IncrementViews(context.Background(), id)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, gallery)
}

func (s *service) Create(ctx context.Context, req GalleryRequest) pkg.Response {
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
	} else if len(req.Description) > 1000 {
		errValidation["description"] = "Description must not exceed 1000 characters"
	}

	// Validate category
	if req.Category == "" {
		errValidation["category"] = "Category is required"
	} else if req.Category != CategoryPhotography && req.Category != CategoryPainting &&
		req.Category != CategorySculpture && req.Category != CategoryDigital &&
		req.Category != CategoryMixed {
		errValidation["category"] = "Invalid category. Must be: photography, painting, sculpture, digital, or mixed"
	}

	// Validate image
	if req.Image == "" {
		errValidation["image"] = "Image URL is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = StatusActive
	}

	// Create gallery
	timeNow := time.Now()
	gallery := &Gallery{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Category:    req.Category,
		Description: req.Description,
		Image:       req.Image,
		Status:      status,
		Views:       0,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	if err := s.repo.Create(ctx, gallery); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create gallery", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Gallery successfully created", nil, gallery)
}

func (s *service) Update(ctx context.Context, id string, req UpdateGalleryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}

	// Check if gallery exists
	_, err := s.repo.FetchGalleryByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
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
		} else if len(req.Description) > 1000 {
			errValidation["description"] = "Description must not exceed 1000 characters"
		} else {
			updateData["description"] = req.Description
		}
	}

	// Validate and set category
	if req.Category != "" {
		if req.Category != CategoryPhotography && req.Category != CategoryPainting &&
			req.Category != CategorySculpture && req.Category != CategoryDigital &&
			req.Category != CategoryMixed {
			errValidation["category"] = "Invalid category. Must be: photography, painting, sculpture, digital, or mixed"
		} else {
			updateData["category"] = req.Category
		}
	}

	// Validate and set status
	if req.Status != "" {
		if req.Status != StatusActive && req.Status != StatusInactive && req.Status != StatusArchived {
			errValidation["status"] = "Invalid status. Must be: active, inactive, or archived"
		} else {
			updateData["status"] = req.Status
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
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update gallery", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Gallery updated successfully", nil, nil)
}

func (s *service) Delete(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}

	// Check if gallery exists
	_, err := s.repo.FetchGalleryByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete gallery", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Gallery deleted successfully", nil, nil)
}
