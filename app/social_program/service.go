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
	GetSocialProgramList(ctx context.Context, params SocialProgramQueryParams, isAdmin bool, accountID string) pkg.Response
	GetSocialProgramBySlug(ctx context.Context, socialProgramSlug string, accountID string) pkg.Response
	GetSocialProgramByID(ctx context.Context, socialProgramID string, accountID string) pkg.Response
	CreateSocialProgram(ctx context.Context, payload SocialProgramRequest) pkg.Response
	UpdateSocialProgram(ctx context.Context, socialProgramID string, payload SocialProgramRequest) pkg.Response
	DeleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response
	ActivateSocialProgram(ctx context.Context, socialProgramID string) pkg.Response
	RejectSocialProgram(ctx context.Context, socialProgramID string, payload RejectSocialProgramRequest) pkg.Response
	CompleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response
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

func (s *service) GetSocialProgramList(ctx context.Context, params SocialProgramQueryParams, isAdmin bool, accountID string) pkg.Response {
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

	if accountID != "" {
		options["account_id"] = accountID
	}

	socialPrograms, err := s.repo.FindAllSocialPrograms(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to fetch social programs")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(socialPrograms) > params.Limit
		hasNext = true
		if len(socialPrograms) > params.Limit {
			socialPrograms = socialPrograms[:params.Limit]
		}
		for i, j := 0, len(socialPrograms)-1; i < j; i, j = i+1, j-1 {
			socialPrograms[i], socialPrograms[j] = socialPrograms[j], socialPrograms[i]
		}
	} else {
		hasNext = len(socialPrograms) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			socialPrograms = socialPrograms[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
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

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toSocialProgramListResponse(socialPrograms, pagination))
}

func (s *service) GetSocialProgramBySlug(ctx context.Context, socialProgramSlug string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	options := map[string]interface{}{
		"slug": socialProgramSlug,
	}
	if accountID != "" {
		options["account_id"] = accountID
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, options)
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

func (s *service) GetSocialProgramByID(ctx context.Context, socialProgramID string, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid social program ID format", nil, nil)
	}

	options := map[string]interface{}{
		"id": socialProgramID,
	}
	if accountID != "" {
		options["account_id"] = accountID
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, options)
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

	errValidation := make(map[string]string)
	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	} else if len(payload.Title) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(payload.Title) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if payload.Description == "" {
		errValidation["description"] = "Deskripsi wajib diisi"
	} else if len(payload.Description) < 10 {
		errValidation["description"] = "Deskripsi minimal 10 karakter"
	} else if len(payload.Description) > 2000 {
		errValidation["description"] = "Deskripsi maksimal 2000 karakter"
	}

	if payload.MinimumAmount <= 0 {
		errValidation["minimumAmount"] = "Minimum donasi harus lebih besar dari 0"
	}

	if payload.BillingDay < 1 || payload.BillingDay > 31 {
		errValidation["billingDay"] = "Hari penagihan harus antara 1 dan 31"
	}

	if payload.CoverImage == nil {
		errValidation["coverImage"] = "Gambar sampul wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	existing, _ := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"title": payload.Title})
	if existing != nil {
		errValidation["title"] = "Program sosial dengan judul ini sudah ada"
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var coverImageURL string
	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "social-programs/covers")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "social_program.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar cover", nil, nil)
		}
		coverImageURL = uploadedURL
	}

	now := time.Now()
	socialProgram := &SocialProgram{
		ID:            uuid.New(),
		Slug:          pkg.Slugify(payload.Title),
		Title:         payload.Title,
		Description:   payload.Description,
		CoverImage:    coverImageURL,
		Status:        StatusPending,
		MinimumAmount: payload.MinimumAmount,
		BillingDay:    payload.BillingDay,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateSocialProgram(ctx, socialProgram); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
			"title":     payload.Title,
		}).WithError(err).Error("failed to create social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat program sosial", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "social_program", socialProgram.ID.String(), nil, socialProgram.toSocialProgramResponse())
	return pkg.NewResponse(http.StatusCreated, "Program sosial berhasil dibuat", nil, socialProgram.toSocialProgramResponse())
}

func (s *service) UpdateSocialProgram(ctx context.Context, id string, payload SocialProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program sosial tidak valid"}, nil)
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	if socialProgram.Status == StatusRejected || socialProgram.Status == StatusCompleted {
		return pkg.NewResponse(http.StatusBadRequest, "Program sosial yang telah ditolak atau selesai tidak dapat diperbarui", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	isActive := socialProgram.Status == StatusActive

	finalTitle := socialProgram.Title
	if payload.Title != "" {
		if !isActive {
			if payload.Title != socialProgram.Title {
				existing, _ := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"title": payload.Title})
				if existing != nil {
					errValidation["title"] = "Program sosial dengan judul ini sudah ada"
				} else {
					finalTitle = payload.Title
					updateData["title"] = payload.Title
					updateData["slug"] = pkg.Slugify(payload.Title)
				}
			}
		} else if payload.Title != socialProgram.Title {
			errValidation["title"] = "Judul tidak dapat diperbarui saat program sosial aktif"
		}
	}

	finalDescription := socialProgram.Description
	if payload.Description != "" {
		finalDescription = payload.Description
		updateData["description"] = payload.Description
	}

	finalMinimumAmount := socialProgram.MinimumAmount
	if payload.MinimumAmount > 0 {
		finalMinimumAmount = payload.MinimumAmount
		updateData["minimum_amount"] = payload.MinimumAmount
	}

	finalBillingDay := socialProgram.BillingDay
	if payload.BillingDay != 0 {
		finalBillingDay = payload.BillingDay
		updateData["billing_day"] = payload.BillingDay
	}

	if len(finalTitle) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(finalTitle) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if finalDescription == "" {
		errValidation["description"] = "Deskripsi wajib diisi"
	} else if len(finalDescription) < 10 {
		errValidation["description"] = "Deskripsi minimal 10 karakter"
	} else if len(finalDescription) > 2000 {
		errValidation["description"] = "Deskripsi maksimal 2000 karakter"
	}

	if finalMinimumAmount <= 0 {
		errValidation["minimumAmount"] = "Minimum donasi harus lebih besar dari 0"
	}

	if finalBillingDay < 1 || finalBillingDay > 31 {
		errValidation["billingDay"] = "Hari penagihan harus antara 1 dan 31"
	}

	if payload.CoverImage == nil && socialProgram.CoverImage == "" {
		errValidation["coverImage"] = "Gambar sampul wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "social-programs/covers")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "social_program.service",
				"id":        id,
			}).WithError(err).Error("failed to upload new cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar cover baru", nil, nil)
		}

		if socialProgram.CoverImage != "" {
			existingImage := s3_pkg.ExtractObjectNameFromURL(socialProgram.CoverImage)
			_ = s.s3Client.DeleteFile(ctx, existingImage)
		}

		updateData["cover_image"] = uploadedURL
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"updateData": "Tidak ada data untuk diperbarui"}, nil)
	}

	updateData["updated_at"] = time.Now()

	if err := s.repo.UpdateSocialProgram(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
			"id":        id,
		}).WithError(err).Error("failed to update social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui program sosial", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "social_program", id, socialProgram.toSocialProgramResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Program sosial berhasil diperbarui", nil, nil)
}

func (s *service) DeleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program sosial tidak valid"}, nil)
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	if socialProgram.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Program sosial aktif, selesai, atau ditolak dan tidak dapat dihapus", nil, nil)
	}

	if err := s.repo.DeleteSocialProgram(ctx, socialProgramID); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
		}).WithError(err).Error("failed to delete social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus program sosial", nil, nil)
	}

	if socialProgram.CoverImage != "" {
		existingImage := s3_pkg.ExtractObjectNameFromURL(socialProgram.CoverImage)
		_ = s.s3Client.DeleteFile(ctx, existingImage)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "social_program", socialProgramID, socialProgram.toSocialProgramResponse(), nil)
	return pkg.NewResponse(http.StatusOK, "Program sosial berhasil dihapus", nil, nil)
}

func (s *service) ActivateSocialProgram(ctx context.Context, socialProgramID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program sosial tidak valid"}, nil)
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	if socialProgram.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya program sosial dengan status pending yang dapat diaktifkan", nil, nil)
	}

	updates := map[string]interface{}{
		"status":     StatusActive,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateSocialProgram(ctx, socialProgramID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
			"id":        socialProgramID,
		}).WithError(err).Error("failed to activate social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengaktifkan program sosial", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "social_program", socialProgramID, socialProgram.toSocialProgramResponse(), updates)
	return pkg.NewResponse(http.StatusOK, "Program sosial berhasil diaktifkan", nil, nil)
}

func (s *service) RejectSocialProgram(ctx context.Context, socialProgramID string, payload RejectSocialProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program sosial tidak valid"}, nil)
	}

	if payload.Reason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"reason": "Alasan penolakan wajib diisi"}, nil)
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	if socialProgram.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya program sosial dengan status pending yang dapat ditolak", nil, nil)
	}

	updates := map[string]interface{}{
		"status":           StatusRejected,
		"rejection_reason": payload.Reason,
		"updated_at":       time.Now(),
	}

	if err := s.repo.UpdateSocialProgram(ctx, socialProgramID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
			"id":        socialProgramID,
		}).WithError(err).Error("failed to reject social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menolak program sosial", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "social_program", socialProgramID, socialProgram.toSocialProgramResponse(), updates)
	return pkg.NewResponse(http.StatusOK, "Program sosial berhasil ditolak", nil, nil)
}

func (s *service) CompleteSocialProgram(ctx context.Context, socialProgramID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(socialProgramID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID program sosial tidak valid"}, nil)
	}

	socialProgram, err := s.repo.FindOneSocialProgram(ctx, map[string]interface{}{"id": socialProgramID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Program sosial tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data program sosial", nil, nil)
	}

	if socialProgram.Status != StatusActive {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya program sosial dengan status aktif yang dapat diselesaikan", nil, nil)
	}

	updates := map[string]interface{}{
		"status":     StatusCompleted,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateSocialProgram(ctx, socialProgramID, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "social_program.service",
			"id":        socialProgramID,
		}).WithError(err).Error("failed to complete social program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyelesaikan program sosial", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "social_program", socialProgramID, socialProgram.toSocialProgramResponse(), updates)
	return pkg.NewResponse(http.StatusOK, "Program sosial berhasil diselesaikan", nil, nil)
}
