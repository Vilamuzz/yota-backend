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
	GetDonationProgramBySlug(ctx context.Context, slug string) pkg.Response
	GetDonationProgramByID(ctx context.Context, donationProgramID string) pkg.Response
	CreateDonationProgram(ctx context.Context, donation DonationProgramRequest) pkg.Response
	UpdateDonationProgram(ctx context.Context, donationProgramID string, payload DonationProgramRequest) pkg.Response
	DeleteDonationProgram(ctx context.Context, donationProgramID string) pkg.Response
	UpdateActiveDonationProgram(ctx context.Context, donationProgramID string) pkg.Response
	UpdateArchivedDonationProgram(ctx context.Context, donationProgramID string) pkg.Response
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
	if params.Page <= 0 {
		params.Page = 1
	}

	options := map[string]interface{}{
		"limit": params.Limit,
		"page":  params.Page,
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.Category != "" {
		options["category"] = params.Category
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}

	if !isAdmin {
		if params.Status != "" {
			options["status"] = params.Status
		} else {
			options["status"] = []string{string(StatusActive), string(StatusExpired), string(StatusCompleted)}
		}
	} else if params.Status != "" {
		options["status"] = params.Status
	}

	total, err := s.repo.CountDonationPrograms(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data donasi", nil, nil)
	}

	donations, err := s.repo.FindAllDonationPrograms(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data donasi", nil, nil)
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit != 0 {
		totalPages++
	}

	pagination := pkg.OffsetPagination{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	if isAdmin {
		return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toAdminDonationProgramListResponse(donations, pagination))
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toDonationProgramListResponse(donations, pagination))
}

func (s *service) GetDonationProgramBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, donation.toDonationProgramResponse())
}

func (s *service) GetDonationProgramByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID donasi tidak valid"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, donation.toAdminDonationProgramResponse())
}

func (s *service) CreateDonationProgram(ctx context.Context, payload DonationProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	status := StatusDraft
	if payload.Status != "" {
		if !payload.Status.IsValid() {
			errValidation["status"] = "Status tidak valid"
		} else {
			status = payload.Status
		}
	}

	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	} else if len(payload.Title) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(payload.Title) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if status == StatusActive {
		if payload.Description == "" {
			errValidation["description"] = "Deskripsi wajib diisi"
		} else if len(payload.Description) < 10 {
			errValidation["description"] = "Deskripsi minimal 10 karakter"
		} else if len(payload.Description) > 2000 {
			errValidation["description"] = "Deskripsi maksimal 2000 karakter"
		}

		if payload.Category == "" {
			errValidation["category"] = "Kategori wajib diisi"
		} else if !payload.Category.IsValid() {
			errValidation["category"] = "Kategori tidak valid"
		}

		if payload.FundTarget <= 0 {
			errValidation["fundTarget"] = "Target dana harus lebih besar dari 0"
		}

		if payload.EndDate == "" {
			errValidation["endDate"] = "Tanggal berakhir wajib diisi"
		}

		if payload.CoverImage == nil {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}
	} else {
		if payload.Description != "" {
			if len(payload.Description) < 10 {
				errValidation["description"] = "Deskripsi minimal 10 karakter"
			} else if len(payload.Description) > 2000 {
				errValidation["description"] = "Deskripsi maksimal 2000 karakter"
			}
		}

		if payload.Category != "" && !payload.Category.IsValid() {
			errValidation["category"] = "Kategori tidak valid"
		}

		if payload.FundTarget < 0 {
			errValidation["fundTarget"] = "Target dana tidak boleh negatif"
		}
	}

	now := time.Now()
	startDate := now
	if payload.StartDate != "" {
		if s, err := time.Parse("2006-01-02", payload.StartDate); err == nil {
			startDate = s
		} else {
			errValidation["startDate"] = "Format tanggal mulai tidak valid (gunakan YYYY-MM-DD)"
		}
	}

	endDate := time.Time{}
	if payload.EndDate != "" {
		if e, err := time.Parse("2006-01-02", payload.EndDate); err == nil {
			endDate = e
		} else {
			errValidation["endDate"] = "Format tanggal berakhir tidak valid (gunakan YYYY-MM-DD)"
		}
	}

	if !endDate.IsZero() && endDate.Before(startDate) {
		errValidation["endDate"] = "Tanggal berakhir harus setelah tanggal mulai"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	existing, _ := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"title": payload.Title})
	if existing != nil {
		errValidation["title"] = "Program donasi dengan judul ini sudah ada"
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var coverImageURL string
	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "donation-programs")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "donation.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar cover", nil, nil)
		}
		coverImageURL = uploadedURL
	}

	donationProgram := &DonationProgram{
		ID:          uuid.New(),
		Title:       payload.Title,
		Slug:        pkg.Slugify(payload.Title),
		Description: payload.Description,
		CoverImage:  coverImageURL,
		Category:    payload.Category,
		FundTarget:  payload.FundTarget,
		Status:      status,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateDonationProgram(ctx, donationProgram); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "donation_program.service",
			"title":     payload.Title,
		}).WithError(err).Error("failed to create donation_program")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat donasi_program", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "donation_program", donationProgram.ID.String(), nil, donationProgram.toAdminDonationProgramResponse())
	return pkg.NewResponse(http.StatusCreated, "Donasi program berhasil dibuat", nil, donationProgram.toAdminDonationProgramResponse())
}

func (s *service) UpdateDonationProgram(ctx context.Context, id string, payload DonationProgramRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID donasi tidak valid"}, nil)
	}

	donationProgram, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}
	if donationProgram.Status == StatusCompleted || donationProgram.Status == StatusExpired || donationProgram.Status == StatusArchived {
		return pkg.NewResponse(http.StatusBadRequest, "Donasi sudah selesai, kedaluwarsa, atau diarsipkan dan tidak dapat diperbarui", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	isActive := donationProgram.Status == StatusActive

	targetStatus := donationProgram.Status
	if payload.Status != "" {
		switch payload.Status {
		case StatusDraft, StatusActive:
			if !isActive {
				targetStatus = payload.Status
				updateData["status"] = payload.Status
			} else if payload.Status != donationProgram.Status {
				errValidation["status"] = "Status donasi aktif tidak dapat dikembalikan ke draft"
			}
		default:
			errValidation["status"] = "Status tidak valid. Hanya draft atau aktif yang dapat diatur."
		}
	}

	finalTitle := donationProgram.Title
	if payload.Title != "" {
		if !isActive {
			if payload.Title != donationProgram.Title {
				existing, _ := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"title": payload.Title})
				if existing != nil {
					errValidation["title"] = "Program donasi dengan judul ini sudah ada"
				} else {
					finalTitle = payload.Title
					updateData["title"] = payload.Title
				}
			}
		} else if payload.Title != donationProgram.Title {
			errValidation["title"] = "Judul tidak dapat diperbarui saat donasi aktif"
		}
	}

	finalDescription := donationProgram.Description
	if payload.Description != "" {
		finalDescription = payload.Description
		updateData["description"] = payload.Description
	}

	finalCategory := donationProgram.Category
	if payload.Category != "" {
		finalCategory = payload.Category
		updateData["category"] = payload.Category
	}

	finalFundTarget := donationProgram.FundTarget
	if payload.FundTarget > 0 {
		if !isActive {
			finalFundTarget = payload.FundTarget
			updateData["fund_target"] = payload.FundTarget
		} else if payload.FundTarget != donationProgram.FundTarget {
			errValidation["fundTarget"] = "Target dana tidak dapat diperbarui saat donasi aktif"
		}
	}

	finalStartDate := donationProgram.StartDate
	if payload.StartDate != "" {
		if s, err := time.Parse("2006-01-02", payload.StartDate); err == nil {
			if !isActive {
				finalStartDate = s
				updateData["start_date"] = s
			} else if !s.Equal(donationProgram.StartDate) {
				errValidation["startDate"] = "Tanggal mulai tidak dapat diperbarui saat donasi aktif"
			}
		} else {
			errValidation["startDate"] = "Format tanggal mulai tidak valid (gunakan YYYY-MM-DD)"
		}
	}

	finalEndDate := donationProgram.EndDate
	if payload.EndDate != "" {
		if e, err := time.Parse("2006-01-02", payload.EndDate); err == nil {
			finalEndDate = e
			updateData["end_date"] = e
		} else {
			errValidation["endDate"] = "Format tanggal berakhir tidak valid (gunakan YYYY-MM-DD)"
		}
	}

	if len(finalTitle) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(finalTitle) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if targetStatus == StatusActive {
		if finalDescription == "" {
			errValidation["description"] = "Deskripsi wajib diisi"
		} else if len(finalDescription) < 10 {
			errValidation["description"] = "Deskripsi minimal 10 karakter"
		} else if len(finalDescription) > 2000 {
			errValidation["description"] = "Deskripsi maksimal 2000 karakter"
		}

		if finalCategory == "" {
			errValidation["category"] = "Kategori wajib diisi"
		} else if !finalCategory.IsValid() {
			errValidation["category"] = "Kategori tidak valid"
		}

		if finalFundTarget <= 0 {
			errValidation["fundTarget"] = "Target dana harus lebih besar dari 0"
		}

		if finalEndDate.IsZero() {
			errValidation["endDate"] = "Tanggal berakhir wajib diisi"
		}

		if payload.CoverImage == nil && donationProgram.CoverImage == "" {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}

		if time.Now().After(finalEndDate) {
			errValidation["status"] = "Tidak dapat mengaktifkan donasi yang telah berakhir"
		}
	} else {
		if finalDescription != "" {
			if len(finalDescription) < 10 {
				errValidation["description"] = "Deskripsi minimal 10 karakter"
			} else if len(finalDescription) > 2000 {
				errValidation["description"] = "Deskripsi maksimal 2000 karakter"
			}
		}

		if finalCategory != "" && !finalCategory.IsValid() {
			errValidation["category"] = "Kategori tidak valid"
		}

		if finalFundTarget < 0 {
			errValidation["fundTarget"] = "Target dana tidak boleh negatif"
		}
	}

	if !finalEndDate.IsZero() && finalEndDate.Before(finalStartDate) {
		errValidation["endDate"] = "Tanggal berakhir harus setelah tanggal mulai"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "donation-programs")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":   "donation.service",
				"donation_id": id,
			}).WithError(err).Error("failed to upload new cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar cover baru", nil, nil)
		}

		if donationProgram.CoverImage != "" {
			existingDonationImage := s3_pkg.ExtractObjectNameFromURL(donationProgram.CoverImage)
			_ = s.s3Client.DeleteFile(ctx, existingDonationImage)
		}

		updateData["cover_image"] = uploadedURL
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"updateData": "Tidak ada data untuk diperbarui"}, nil)
	}

	updateData["updated_at"] = time.Now()

	if err := s.repo.UpdateDonationProgram(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to update donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui donasi", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "donation", id, donationProgram.toAdminDonationProgramResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Donasi berhasil diperbarui", nil, nil)
}

func (s *service) DeleteDonationProgram(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID donasi tidak valid"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}

	if donation.Status != StatusDraft {
		return pkg.NewResponse(http.StatusBadRequest, "Donasi aktif, selesai, atau kedaluwarsa dan tidak dapat dihapus", nil, nil)
	}

	if err := s.repo.DeleteDonationProgram(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to delete donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus donasi", nil, nil)
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
	return pkg.NewResponse(http.StatusOK, "Donasi berhasil dihapus", nil, nil)
}

func (s *service) UpdateActiveDonationProgram(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID donasi tidak valid"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}

	if donation.Status == StatusActive {
		return pkg.NewResponse(http.StatusOK, "Donasi sudah aktif", nil, nil)
	}

	if donation.Status != StatusDraft {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya donasi draft yang dapat diaktifkan", nil, nil)
	}

	errValidation := make(map[string]string)
	if donation.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	}
	if donation.Description == "" {
		errValidation["description"] = "Deskripsi wajib diisi"
	}
	if donation.Category == "" {
		errValidation["category"] = "Kategori wajib diisi"
	}
	if donation.FundTarget <= 0 {
		errValidation["fundTarget"] = "Target dana harus lebih besar dari 0"
	}
	if donation.EndDate.IsZero() {
		errValidation["endDate"] = "Tanggal berakhir wajib diisi"
	} else if time.Now().After(donation.EndDate) {
		errValidation["endDate"] = "Tidak dapat mengaktifkan donasi yang telah berakhir"
	}
	if donation.CoverImage == "" {
		errValidation["coverImage"] = "Gambar sampul wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusActive,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateDonationProgram(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to activate donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengaktifkan donasi", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "donation", id, donation.toAdminDonationProgramResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Donasi berhasil diaktifkan", nil, nil)
}

func (s *service) UpdateArchivedDonationProgram(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID donasi tidak valid"}, nil)
	}

	donation, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}

	if donation.Status == StatusArchived {
		return pkg.NewResponse(http.StatusOK, "Donasi sudah diarsipkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusArchived,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateDonationProgram(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "donation.service",
			"donation_id": id,
		}).WithError(err).Error("failed to archive donation")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengarsipkan donasi", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "donation", id, donation.toAdminDonationProgramResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Donasi berhasil diarsipkan", nil, nil)
}

func (s *service) UpdateExpiredDonationProgram(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repo.UpdateExpiredDonationProgram(ctx)
}
