package donation_program

import (
	"context"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetDonationProgramList(ctx context.Context, params DonationProgramQueryParams, isAdmin bool) pkg.Response
	GetPublishedDonationProgramBySlug(ctx context.Context, slug string) pkg.Response
	GetDonationProgramByID(ctx context.Context, donationProgramID string) pkg.Response
	CreateDonationProgram(ctx context.Context, donation DonationProgramRequest) pkg.Response
	UpdateDonationProgram(ctx context.Context, donationProgramID string, payload DonationProgramRequest) pkg.Response
	DeleteDonationProgram(ctx context.Context, donationProgramID string) pkg.Response
	UpdateExpiredDonationProgram(ctx context.Context) error
}

type service struct {
	repo       Repository
	logService app_log.Service
	s3Client   s3_pkg.Client
	timeout    time.Duration
}

func NewService(repo Repository, logService app_log.Service, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:       repo,
		logService: logService,
		s3Client:   s3Client,
		timeout:    timeout,
	}
}

func (s *service) GetDonationProgramList(ctx context.Context, params DonationProgramQueryParams, isAdmin bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.Category != "" {
		options["category"] = params.Category
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	if isAdmin {
		if params.Status != "" {
			options["status"] = params.Status
		}
	} else {
		options["published"] = true
	}

	donations, err := s.repo.FindAllDonationPrograms(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch donations", nil, nil)
	}

	hasNext := len(donations) > params.Limit
	if hasNext {
		donations = donations[:params.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := params.PrevCursor != ""
	if hasNext && len(donations) > 0 {
		lastDonation := donations[len(donations)-1]
		nextCursor = pkg.EncodeCursor(lastDonation.CreatedAt, lastDonation.ID.String())
	}
	if hasPrev && len(donations) > 0 {
		firstDonation := donations[0]
		prevCursor = pkg.EncodeCursor(firstDonation.CreatedAt, firstDonation.ID.String())
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	if isAdmin {
		return pkg.NewResponse(http.StatusOK, "Success", nil, toDonationProgramListResponse(donations, pagination))
	}
	return pkg.NewResponse(http.StatusOK, "Success", nil, toPublishedDonationProgramListResponse(donations, pagination))
}

func (s *service) GetPublishedDonationProgramBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation.toPublishedDonationProgramResponse())
}

func (s *service) GetDonationProgramByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, donation)
}

func (s *service) CreateDonationProgram(ctx context.Context, payload DonationProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	status := StatusDraft
	if payload.Status != "" {
		if !payload.Status.IsValid() {
			errValidation["status"] = "Invalid status"
		} else {
			status = payload.Status
		}
	}

	if payload.Title == "" {
		errValidation["title"] = "Title is required"
	} else if len(payload.Title) < 3 {
		errValidation["title"] = "Title must be at least 3 characters"
	} else if len(payload.Title) > 200 {
		errValidation["title"] = "Title must not exceed 200 characters"
	}

	if status == StatusActive {
		if payload.Description == "" {
			errValidation["description"] = "Description is required"
		} else if len(payload.Description) < 10 {
			errValidation["description"] = "Description must be at least 10 characters"
		} else if len(payload.Description) > 2000 {
			errValidation["description"] = "Description must not exceed 2000 characters"
		}

		if payload.Category == "" {
			errValidation["category"] = "Category is required"
		} else if !payload.Category.IsValid() {
			errValidation["category"] = "Invalid category"
		}

		if payload.FundTarget <= 0 {
			errValidation["fund_target"] = "Fund target must be greater than 0"
		}

		if payload.EndDate.IsZero() {
			errValidation["end_date"] = "End date is required"
		}

		if payload.CoverImage == nil {
			errValidation["coverImage"] = "Cover image is required"
		}
	} else {
		// Draft mode validations
		if payload.Description != "" {
			if len(payload.Description) < 10 {
				errValidation["description"] = "Description must be at least 10 characters"
			} else if len(payload.Description) > 2000 {
				errValidation["description"] = "Description must not exceed 2000 characters"
			}
		}

		if payload.Category != "" && !payload.Category.IsValid() {
			errValidation["category"] = "Invalid category"
		}

		if payload.FundTarget < 0 {
			errValidation["fund_target"] = "Fund target must not be negative"
		}
	}

	now := time.Now()
	startDate := now
	if !payload.StartDate.IsZero() {
		startDate = payload.StartDate
	}

	if !payload.EndDate.IsZero() && payload.EndDate.Before(startDate) {
		errValidation["end_date"] = "End date must be after start date"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var coverImageURL string
	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		coverImageURL = uploadedURL
	}

	donation := &DonationProgram{
		ID:          uuid.New(),
		Title:       payload.Title,
		Slug:        pkg.Slugify(payload.Title),
		Description: payload.Description,
		CoverImage:  coverImageURL,
		Category:    payload.Category,
		FundTarget:  payload.FundTarget,
		Status:      status,
		StartDate:   startDate,
		EndDate:     payload.EndDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateDonationProgram(ctx, donation); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation.service",
			"title":     payload.Title,
		}).WithError(err).Error("failed to create donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create donation", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "donation", donation.ID.String(), nil, donation.toDonationProgramResponse())
	return pkg.NewResponse(http.StatusCreated, "Donation successfully created", nil, donation.toDonationProgramResponse())
}

func (s *service) UpdateDonationProgram(ctx context.Context, id string, req DonationProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donationProgram, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}
	if donationProgram.Status == StatusCompleted || donationProgram.Status == StatusExpired {
		return pkg.NewResponse(http.StatusBadRequest, "Donation is completed or expired and cannot be updated", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	isActive := donationProgram.Status == StatusActive

	targetStatus := donationProgram.Status
	if req.Status != "" {
		if !req.Status.IsValid() {
			errValidation["status"] = "Invalid status"
		} else if !isActive || req.Status == StatusArchived {
			targetStatus = req.Status
			updateData["status"] = req.Status
		} else if req.Status != donationProgram.Status {
			errValidation["status"] = "Status cannot be updated when donation is active"
		}
	}

	finalTitle := donationProgram.Title
	if req.Title != "" {
		if !isActive {
			finalTitle = req.Title
			updateData["title"] = req.Title
		} else if req.Title != donationProgram.Title {
			errValidation["title"] = "Title cannot be updated when donation is active"
		}
	}

	finalDescription := donationProgram.Description
	if req.Description != "" {
		finalDescription = req.Description
		updateData["description"] = req.Description
	}

	finalCategory := donationProgram.Category
	if req.Category != "" {
		finalCategory = req.Category
		updateData["category"] = req.Category
	}

	finalFundTarget := donationProgram.FundTarget
	if req.FundTarget > 0 {
		if !isActive {
			finalFundTarget = req.FundTarget
			updateData["fund_target"] = req.FundTarget
		} else if req.FundTarget != donationProgram.FundTarget {
			errValidation["fund_target"] = "Fund target cannot be updated when donation is active"
		}
	}

	finalStartDate := donationProgram.StartDate
	if !req.StartDate.IsZero() {
		if !isActive {
			finalStartDate = req.StartDate
			updateData["start_date"] = req.StartDate
		} else if !req.StartDate.Equal(donationProgram.StartDate) {
			errValidation["start_date"] = "Start date cannot be updated when donation is active"
		}
	}

	finalEndDate := donationProgram.EndDate
	if !req.EndDate.IsZero() {
		finalEndDate = req.EndDate
		updateData["end_date"] = req.EndDate
	}

	if len(finalTitle) < 3 {
		errValidation["title"] = "Title must be at least 3 characters"
	} else if len(finalTitle) > 200 {
		errValidation["title"] = "Title must not exceed 200 characters"
	}

	if targetStatus == StatusActive {
		if finalDescription == "" {
			errValidation["description"] = "Description is required"
		} else if len(finalDescription) < 10 {
			errValidation["description"] = "Description must be at least 10 characters"
		} else if len(finalDescription) > 2000 {
			errValidation["description"] = "Description must not exceed 2000 characters"
		}

		if finalCategory == "" {
			errValidation["category"] = "Category is required"
		} else if !finalCategory.IsValid() {
			errValidation["category"] = "Invalid category"
		}

		if finalFundTarget <= 0 {
			errValidation["fund_target"] = "Fund target must be greater than 0"
		}

		if finalEndDate.IsZero() {
			errValidation["end_date"] = "End date is required"
		}

		if req.CoverImage == nil && donationProgram.CoverImage == "" {
			errValidation["coverImage"] = "Cover image is required"
		}

		if time.Now().After(finalEndDate) {
			errValidation["status"] = "Cannot activate donation that has already ended"
		}
	} else {
		if finalDescription != "" {
			if len(finalDescription) < 10 {
				errValidation["description"] = "Description must be at least 10 characters"
			} else if len(finalDescription) > 2000 {
				errValidation["description"] = "Description must not exceed 2000 characters"
			}
		}

		if finalCategory != "" && !finalCategory.IsValid() {
			errValidation["category"] = "Invalid category"
		}

		if finalFundTarget < 0 {
			errValidation["fund_target"] = "Fund target must not be negative"
		}
	}

	if !finalEndDate.IsZero() && finalEndDate.Before(finalStartDate) {
		errValidation["end_date"] = "End date must be after start date"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if req.CoverImage != nil {
		existingDonationImage := s3_pkg.ExtractObjectNameFromURL(donationProgram.CoverImage)
		if err := s.s3Client.DeleteFile(ctx, existingDonationImage); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete existing image", nil, nil)
		}
		uploadedURL, err := s.s3Client.UploadFile(ctx, req.CoverImage, "donations")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		updateData["cover_image"] = uploadedURL
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	updateData["updated_at"] = time.Now()

	if err := s.repo.UpdateDonationProgram(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to update donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update donation", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "donation", id, donationProgram.toDonationProgramResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Donation updated successfully", nil, nil)
}

func (s *service) DeleteDonationProgram(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid donation ID format"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donation not found", nil, nil)
	}

	if donation.Status != StatusDraft {
		return pkg.NewResponse(http.StatusBadRequest, "Donation is active, completed, or expired and cannot be deleted", nil, nil)
	}

	if err := s.repo.DeleteDonationProgram(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to delete donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete donation", nil, nil)
	}

	if donation.CoverImage != "" {
		imageObjectName := s3_pkg.ExtractObjectNameFromURL(donation.CoverImage)
		if err := s.s3Client.DeleteFile(ctx, imageObjectName); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":   "donation.service",
				"donation_id": id,
			}).WithError(err).Error("failed to delete cover image from S3")
		}
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "donation", id, donation.toDonationProgramResponse(), nil)
	return pkg.NewResponse(http.StatusOK, "Donation deleted successfully", nil, nil)
}

func (s *service) UpdateExpiredDonationProgram(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repo.UpdateExpiredDonationProgram(ctx)
}
