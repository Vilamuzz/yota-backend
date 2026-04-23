package social_program

import (
	"context"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetSocialProgramList(ctx context.Context, params SocialProgramQueryParams, isAdmin bool) pkg.Response
	GetSocialProgramBySlug(ctx context.Context, socialProgramSlug string) pkg.Response
	CreateSocialProgram(ctx context.Context, payload SocialProgramRequest) pkg.Response
	UpdateSocialProgram(ctx context.Context, socialProgramID string, payload SocialProgramRequest) pkg.Response
	DeleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response
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

func (s *service) GetSocialProgramList(ctx context.Context, params SocialProgramQueryParams, isAdmin bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = params.PrevCursor
	}

	if !isAdmin {
		options["status"] = string(StatusActive)
	} else if params.Status != "" {
		options["status"] = params.Status
	}

	if params.Search != "" {
		options["search"] = params.Search
	}

	socialPrograms, err := s.repo.FindAllSocialPrograms(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to fetch social programs")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch social programs", nil, nil)
	}

	hasMore := len(socialPrograms) > params.Limit
	if hasMore {
		socialPrograms = socialPrograms[:params.Limit]
	}

	if usingPrevCursor {
		for i, j := 0, len(socialPrograms)-1; i < j; i, j = i+1, j-1 {
			socialPrograms[i], socialPrograms[j] = socialPrograms[j], socialPrograms[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(socialPrograms) > 0 {
		first := socialPrograms[0]
		last := socialPrograms[len(socialPrograms)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Social programs found successfully", nil, toSocialProgramListResponse(socialPrograms, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetSocialProgramBySlug(ctx context.Context, socialProgramSlug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{
		"slug": socialProgramSlug,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Social program not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to fetch social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch social program", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Social program found successfully", nil, socialProgram.toSocialProgramResponse())
}

func (s *service) CreateSocialProgram(ctx context.Context, payload SocialProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if payload.CoverImage == nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"cover_image": "Cover image is required"}, nil)
	}

	coverImageURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "social-programs/covers")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to upload cover image")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload cover image", nil, nil)
	}

	now := time.Now()
	socialProgram := &SocialProgram{
		ID:            uuid.New(),
		Slug:          pkg.Slugify(payload.Title),
		Title:         payload.Title,
		Description:   payload.Description,
		CoverImage:    coverImageURL,
		Status:        payload.Status,
		MinimumAmount: payload.MinimumAmount,
		BillingDay:    payload.BillingDay,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateSocialProgram(ctx, socialProgram); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to create social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create social program", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Social program created successfully", nil, socialProgram.toSocialProgramResponse())
}

func (s *service) UpdateSocialProgram(ctx context.Context, socialProgramID string, payload SocialProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Social program not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch social program", nil, nil)
	}

	updates := map[string]interface{}{
		"title":          payload.Title,
		"description":    payload.Description,
		"status":         payload.Status,
		"minimum_amount": payload.MinimumAmount,
		"billing_day":    payload.BillingDay,
		"updated_at":     time.Now(),
	}

	if payload.CoverImage != nil {
		coverImageURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "social-programs/covers")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "social_program.service",
			}).WithError(err).Error("failed to upload new cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload cover image", nil, nil)
		}
		updates["cover_image"] = coverImageURL

		if socialProgram.CoverImage != "" {
			_ = s.s3Client.DeleteFile(ctx, socialProgram.CoverImage)
		}
	}

	if err := s.repo.UpdateSocialProgram(ctx, socialProgramID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to update social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update social program", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Social program updated successfully", nil, nil)
}

func (s *service) DeleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Social program not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch social program", nil, nil)
	}

	if err := s.repo.DeleteSocialProgram(ctx, socialProgramID); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to delete social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete social program", nil, nil)
	}

	if socialProgram.CoverImage != "" {
		_ = s.s3Client.DeleteFile(ctx, socialProgram.CoverImage)
	}

	return pkg.NewResponse(http.StatusOK, "Social program deleted successfully", nil, nil)
}
