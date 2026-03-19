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
	ListPublishedDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response
	GetPublishedDonationBySlug(ctx context.Context, slug string) pkg.Response
	ListDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response
	GetDonationByID(ctx context.Context, id string) pkg.Response
	CreateDonation(ctx context.Context, donation DonationRequest) pkg.Response
	UpdateDonation(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response
	DeleteDonation(ctx context.Context, id string) pkg.Response
	UpdateExpiredDonations(ctx context.Context) error
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

func (s *service) ListPublishedDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit":     queryParams.Limit,
		"published": true,
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
	}
	if queryParams.Category != "" {
		options["category"] = queryParams.Category
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	donations, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch donations", nil, nil)
	}

	hasNext := len(donations) > queryParams.Limit
	if hasNext {
		donations = donations[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(donations) > 0 {
		lastDonation := donations[len(donations)-1]
		nextCursor = pkg.EncodeCursor(lastDonation.CreatedAt, lastDonation.ID)
	}
	if hasPrev && len(donations) > 0 {
		firstDonation := donations[0]
		prevCursor = pkg.EncodeCursor(firstDonation.CreatedAt, firstDonation.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toPublishedDonationListResponse(donations, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) ListDonations(ctx context.Context, queryParams DonationQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.Category != "" {
		options["category"] = queryParams.Category
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	donations, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch donations", nil, nil)
	}

	hasNext := len(donations) > queryParams.Limit
	if hasNext {
		donations = donations[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(donations) > 0 {
		lastDonation := donations[len(donations)-1]
		nextCursor = pkg.EncodeCursor(lastDonation.CreatedAt, lastDonation.ID)
	}
	if hasPrev && len(donations) > 0 {
		firstDonation := donations[0]
		prevCursor = pkg.EncodeCursor(firstDonation.CreatedAt, firstDonation.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toDonationListResponse(donations, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) GetPublishedDonationBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	donation, err := s.repo.FindOne(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation.toPublishedDonationResponse())
}

func (s *service) GetDonationByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindOne(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation)
}

func (s *service) CreateDonation(ctx context.Context, req DonationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Title == "" {
		errValidation["title"] = "Title is required"
	} else if len(req.Title) < 3 {
		errValidation["title"] = "Title must be at least 3 characters"
	} else if len(req.Title) > 200 {
		errValidation["title"] = "Title must not exceed 200 characters"
	}

	if req.Description == "" {
		errValidation["description"] = "Description is required"
	} else if len(req.Description) < 10 {
		errValidation["description"] = "Description must be at least 10 characters"
	} else if len(req.Description) > 2000 {
		errValidation["description"] = "Description must not exceed 2000 characters"
	}

	if req.Category == "" {
		errValidation["category"] = "Category is required"
	} else if !req.Category.IsValid() {
		errValidation["category"] = "Invalid category"
	}

	if req.FundTarget <= 0 {
		errValidation["fund_target"] = "Fund target must be greater than 0"
	}

	now := time.Now()
	if req.DateEnd.IsZero() {
		errValidation["date_end"] = "End date is required"
	} else if req.DateEnd.Before(now) {
		errValidation["date_end"] = "End date must be in the future"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var imageURL string
	if req.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.Image, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		imageURL = uploadedURL
	}

	status := StatusDraft
	var publishedAt *time.Time
	if req.Status {
		status = StatusActive
		publishedAt = &now
	}

	donation := &Donation{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Slug:        pkg.Slugify(req.Title),
		Description: req.Description,
		ImageURL:    imageURL,
		Category:    req.Category,
		FundTarget:  req.FundTarget,
		Status:      status,
		DateEnd:     req.DateEnd,
		PublishedAt: publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, donation); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Donation successfully created", nil, donation)
}

func (s *service) UpdateDonation(ctx context.Context, id string, req UpdateDonationRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	existingDonation, err := s.repo.FindOne(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}
	if existingDonation.Status == StatusCompleted || existingDonation.Status == StatusExpired {
		return pkg.NewResponse(http.StatusBadRequest, "Donation is completed or expired and cannot be updated", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})
	if req.Title != "" {
		if len(req.Title) < 3 {
			errValidation["title"] = "Title must be at least 3 characters"
		} else if len(req.Title) > 200 {
			errValidation["title"] = "Title must not exceed 200 characters"
		} else {
			updateData["title"] = req.Title
		}
	}

	if req.Description != "" {
		if len(req.Description) < 10 {
			errValidation["description"] = "Description must be at least 10 characters"
		} else if len(req.Description) > 2000 {
			errValidation["description"] = "Description must not exceed 2000 characters"
		} else {
			updateData["description"] = req.Description
		}
	}

	if req.Category != "" {
		if !req.Category.IsValid() {
			errValidation["category"] = "Invalid category"
		} else {
			updateData["category"] = req.Category
		}
	}

	if req.FundTarget > 0 {
		updateData["fund_target"] = req.FundTarget
	} else if req.FundTarget < 0 {
		errValidation["fund_target"] = "Fund target must be greater than 0"
	}

	if req.Status != nil {
		status := StatusDraft
		if *req.Status {
			status = StatusActive
		}

		if status == StatusActive && time.Now().After(existingDonation.DateEnd) {
			errValidation["status"] = "Cannot activate donation that has already ended"
		} else {
			if status == StatusActive && existingDonation.PublishedAt == nil {
				now := time.Now()
				updateData["published_at"] = &now
			}
			updateData["status"] = status
		}
	}

	if !req.DateEnd.IsZero() {
		if req.DateEnd.Before(time.Now()) {
			errValidation["date_end"] = "End date must be in the future"
		} else {
			updateData["date_end"] = req.DateEnd
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if req.Image != nil {
		existingDonationImage := s3_pkg.ExtractObjectNameFromURL(existingDonation.ImageURL)
		if err := s.s3Client.DeleteFile(ctx, existingDonationImage); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete existing image", nil, nil)
		}
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.Image, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		updateData["image_url"] = uploadedURL
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	updateData["updated_at"] = time.Now()

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update donation", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Donation updated successfully", nil, nil)
}

func (s *service) DeleteDonation(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindOne(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	if donation.Status == StatusCompleted || donation.Status == StatusExpired || donation.Status == StatusActive {
		return pkg.NewResponse(http.StatusBadRequest, "Donation is active, completed, or expired and cannot be deleted", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete donation", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Donation deleted successfully", nil, nil)
}

func (s *service) UpdateExpiredDonations(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repo.UpdateExpired(ctx)
}
