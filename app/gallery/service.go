package gallery

import (
	"context"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetGalleryList(ctx context.Context, params GalleryQueryParams, isAdmin bool) pkg.Response
	GetGalleryBySlug(ctx context.Context, slug string) pkg.Response
	GetGalleryByID(ctx context.Context, id string) pkg.Response
	CreateGallery(ctx context.Context, payload GalleryCreateRequest) pkg.Response
	UpdateGallery(ctx context.Context, id string, payload GalleryUpdateRequest) pkg.Response
	DeleteGallery(ctx context.Context, id string) pkg.Response
	UpdatePublishGallery(ctx context.Context, id string) pkg.Response
	UpdateArchivedGallery(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo         Repository
	logService   app_log.Service
	s3Client     s3_pkg.Client
	mediaService media.Service
	timeout      time.Duration
}

func NewService(repo Repository, logService app_log.Service, s3Client s3_pkg.Client, mediaService media.Service, timeout time.Duration) Service {
	return &service{
		repo:         repo,
		logService:   logService,
		s3Client:     s3Client,
		mediaService: mediaService,
		timeout:      timeout,
	}
}

func (s *service) GetGalleryList(ctx context.Context, params GalleryQueryParams, isAdmin bool) pkg.Response {
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

	if isAdmin {
		if params.Status != "" {
			options["status"] = params.Status
		}
	} else {
		options["status"] = media.MediaStatusPublished
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

	galleries, err := s.repo.FindAllGalleries(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data galeri", nil, nil)
	}

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(galleries) > params.Limit
		hasNext = true
		if len(galleries) > params.Limit {
			galleries = galleries[:params.Limit]
		}
		// Reverse the slice because the repository returns ASC order for PrevCursor
		for i, j := 0, len(galleries)-1; i < j; i, j = i+1, j-1 {
			galleries[i], galleries[j] = galleries[j], galleries[i]
		}
	} else {
		hasNext = len(galleries) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			galleries = galleries[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
	if len(galleries) > 0 {
		first := galleries[0]
		last := galleries[len(galleries)-1]
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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toGalleryListResponse(galleries, pagination))
}

func (s *service) GetGalleryBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	gallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	go s.repo.IncrementViews(context.Background(), gallery.ID.String())

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, gallery.toGalleryResponse())
}

func (s *service) GetGalleryByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}

	gallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, gallery.toGalleryResponse())
}

func (s *service) CreateGallery(ctx context.Context, payload GalleryCreateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	status := media.MediaStatusDraft
	if payload.Status != "" {
		status = payload.Status
	}

	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	} else if len(payload.Title) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(payload.Title) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if status == media.MediaStatusPublished {
		if payload.Description == "" {
			errValidation["description"] = "Deskripsi wajib diisi"
		} else if len(payload.Description) < 10 {
			errValidation["description"] = "Deskripsi minimal 10 karakter"
		} else if len(payload.Description) > 1000 {
			errValidation["description"] = "Deskripsi maksimal 1000 karakter"
		}

		if payload.Category == "" {
			errValidation["category"] = "Kategori wajib diisi"
		}
		if len(payload.MediaFiles) == 0 {
			errValidation["media"] = "Minimal satu file media wajib diisi"
		}
		if payload.CoverImage == nil {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}
	} else {
		if payload.Description != "" {
			if len(payload.Description) < 10 {
				errValidation["description"] = "Deskripsi minimal 10 karakter"
			} else if len(payload.Description) > 1000 {
				errValidation["description"] = "Deskripsi maksimal 1000 karakter"
			}
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	existing, _ := s.repo.FindOneGallery(ctx, map[string]interface{}{"title": payload.Title})
	if existing != nil {
		errValidation["title"] = "Galeri dengan judul ini sudah ada"
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var mediaItems []media.Media
	var coverImageURL string

	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "galleries")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "gallery.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar sampul", nil, nil)
		}
		coverImageURL = uploadedURL
	}

	for i, file := range payload.MediaFiles {
		var finalURL string
		if file != nil {
			url, err := s.s3Client.UploadFile(ctx, file, "galleries")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"title":     payload.Title,
				}).WithError(err).Error("failed to upload media")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file media", nil, nil)
			}
			finalURL = url
		}

		if finalURL != "" {
			alt := ""
			if i < len(payload.MediaAlts) {
				alt = payload.MediaAlts[i]
			}

			item := media.Media{
				ID:    uuid.New(),
				Type:  media.MediaTypeImage, // Default to image
				URL:   finalURL,
				Alt:   alt,
				Order: i + 1,
			}

			if i == 0 && coverImageURL == "" {
				coverImageURL = finalURL
			}

			mediaItems = append(mediaItems, item)
		}
	}

	timeNow := time.Now()

	gallery := &Gallery{
		ID:          uuid.New(),
		Title:       payload.Title,
		Slug:        pkg.Slugify(payload.Title),
		Category:    payload.Category,
		Description: payload.Description,
		CoverImage:  coverImageURL,
		Status:      status,
		Views:       0,
		Media:       mediaItems,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	if err := s.repo.CreateGallery(ctx, gallery); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "gallery.service",
			"title":     payload.Title,
		}).WithError(err).Error("failed to create gallery")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat galeri", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "gallery", gallery.ID.String(), nil, gallery.toGalleryResponse())

	return pkg.NewResponse(http.StatusCreated, "Galeri berhasil dibuat", nil, gallery.toGalleryResponse())
}

func (s *service) UpdateGallery(ctx context.Context, id string, payload GalleryUpdateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}

	existingGallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	targetStatus := existingGallery.Status
	if payload.Status != "" {
		targetStatus = payload.Status
		updateData["status"] = payload.Status
	}

	finalTitle := existingGallery.Title
	if payload.Title != "" {
		if payload.Title != existingGallery.Title {
			existing, _ := s.repo.FindOneGallery(ctx, map[string]interface{}{"title": payload.Title})
			if existing != nil {
				errValidation["title"] = "Galeri dengan judul ini sudah ada"
			} else {
				finalTitle = payload.Title
				updateData["title"] = payload.Title
				updateData["slug"] = pkg.Slugify(payload.Title)
			}
		}
	}

	finalDescription := existingGallery.Description
	if payload.Description != "" {
		finalDescription = payload.Description
		updateData["description"] = payload.Description
	}

	finalCategory := existingGallery.Category
	if payload.Category != "" {
		finalCategory = payload.Category
		updateData["category"] = payload.Category
	}

	if len(finalTitle) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	} else if len(finalTitle) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if targetStatus == media.MediaStatusPublished {
		if finalDescription == "" {
			errValidation["description"] = "Deskripsi wajib diisi"
		} else if len(finalDescription) < 10 {
			errValidation["description"] = "Deskripsi minimal 10 karakter"
		} else if len(finalDescription) > 1000 {
			errValidation["description"] = "Deskripsi maksimal 1000 karakter"
		}

		if finalCategory == "" {
			errValidation["category"] = "Kategori wajib diisi"
		}

		if len(payload.MediaFiles) == 0 && len(existingGallery.Media) == 0 {
			errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
		}

		if payload.CoverImage == nil && existingGallery.CoverImage == "" {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}
	} else {
		if finalDescription != "" {
			if len(finalDescription) < 10 {
				errValidation["description"] = "Deskripsi minimal 10 karakter"
			} else if len(finalDescription) > 1000 {
				errValidation["description"] = "Deskripsi maksimal 1000 karakter"
			}
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// Handle Media Updates (Deletions, Updates, and Additions)
	existingMediaList, err := s.mediaService.FetchEntityMedia(ctx, id, "galleries")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data media yang ada", nil, nil)
	}

	// Validate update slices lengths
	if len(payload.UpdateMediaAlts) > 0 && len(payload.UpdateMediaAlts) != len(payload.MediaIDs) {
		errValidation["updateMediaAlts"] = "Jumlah updateMediaAlts harus sama dengan jumlah mediaIds"
	}
	if len(payload.UpdateMediaOrders) > 0 && len(payload.UpdateMediaOrders) != len(payload.MediaIDs) {
		errValidation["updateMediaOrders"] = "Jumlah updateMediaOrders harus sama dengan jumlah mediaIds"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// 1. Identify and Delete Missing Media
	payloadMediaIDMap := make(map[string]bool)
	for _, mid := range payload.MediaIDs {
		payloadMediaIDMap[mid] = true
	}

	for _, em := range existingMediaList {
		if !payloadMediaIDMap[em.ID.String()] {
			if err := s.mediaService.DeleteMediaByID(ctx, em.ID.String()); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"media_id":  em.ID,
				}).WithError(err).Warn("failed to delete media")
			}
		}
	}

	// 2. Update Existing Media Metadata
	for i, mid := range payload.MediaIDs {
		updateMediaData := make(map[string]interface{})
		if i < len(payload.UpdateMediaAlts) {
			updateMediaData["alt"] = payload.UpdateMediaAlts[i]
		}
		if i < len(payload.UpdateMediaOrders) {
			updateMediaData["order"] = payload.UpdateMediaOrders[i]
		}

		if len(updateMediaData) > 0 {
			if err := s.mediaService.UpdateMediaByID(ctx, mid, updateMediaData); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"media_id":  mid,
				}).WithError(err).Warn("failed to update media metadata")
			}
		}
	}

	// 3. Add New Media Files
	for i, file := range payload.MediaFiles {
		if file != nil {
			url, err := s.s3Client.UploadFile(ctx, file, "galleries")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"id":        id,
				}).WithError(err).Error("failed to upload media")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file media", nil, nil)
			}

			alt := ""
			if i < len(payload.MediaAlts) {
				alt = payload.MediaAlts[i]
			}

			order := 0
			if i < len(payload.MediaOrders) {
				order = payload.MediaOrders[i]
			}

			newMedia := []media.Media{
				{
					ID:    uuid.New(),
					URL:   url,
					Type:  media.MediaTypeImage,
					Alt:   alt,
					Order: order,
				},
			}

			if err := s.mediaService.CreateEntityMedia(ctx, id, "galleries", newMedia); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"id":        id,
				}).WithError(err).Error("failed to save new media data")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan data media baru", nil, nil)
			}
		}
	}

	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "galleries")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "gallery.service",
				"gallery_id": id,
			}).WithError(err).Error("failed to upload new cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar sampul baru", nil, nil)
		}

		if existingGallery.CoverImage != "" {
			existingCoverImage := s3_pkg.ExtractObjectNameFromURL(existingGallery.CoverImage)
			_ = s.s3Client.DeleteFile(ctx, existingCoverImage)
		}

		updateData["cover_image"] = uploadedURL
	}

	if len(updateData) == 0 && len(payload.MediaFiles) == 0 && payload.CoverImage == nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"updateData": "Tidak ada data untuk diperbarui"}, nil)
	}

	if len(updateData) > 0 {
		updateData["updated_at"] = time.Now()

		if err := s.repo.UpdateGallery(ctx, id, updateData); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "gallery.service",
				"gallery_id": id,
			}).WithError(err).Error("failed to update gallery")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui galeri", nil, nil)
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "gallery", id, existingGallery.toGalleryResponse(), updateData)

	return pkg.NewResponse(http.StatusOK, "Galeri berhasil diperbarui", nil, nil)
}

func (s *service) DeleteGallery(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}
	existingGallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	if err := s.mediaService.DeleteEntityMedia(ctx, id, "galleries"); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "gallery.service",
			"gallery_id": id,
		}).WithError(err).Warn("failed to delete associated media")
	}

	if existingGallery.CoverImage != "" {
		objectName := s3_pkg.ExtractObjectNameFromURL(existingGallery.CoverImage)
		if objectName != "" {
			_ = s.s3Client.DeleteFile(ctx, objectName)
		}
	}

	if err := s.repo.DeleteGallery(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "gallery.service",
			"gallery_id": id,
		}).WithError(err).Error("failed to delete gallery record")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus galeri", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "gallery", id, existingGallery.toGalleryResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Galeri berhasil dihapus", nil, nil)
}

func (s *service) UpdatePublishGallery(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}

	gallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	if gallery.Status == media.MediaStatusPublished {
		return pkg.NewResponse(http.StatusOK, "Galeri sudah dipublikasikan", nil, nil)
	}

	errValidation := make(map[string]string)
	if len(gallery.Title) < 3 {
		errValidation["title"] = "Judul minimal 3 karakter"
	}

	if gallery.Description == "" {
		errValidation["description"] = "Deskripsi wajib diisi"
	} else if len(gallery.Description) < 10 {
		errValidation["description"] = "Deskripsi minimal 10 karakter"
	}

	if gallery.Category == "" {
		errValidation["category"] = "Kategori wajib diisi"
	}

	if len(gallery.Media) == 0 {
		errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
	}

	if gallery.CoverImage == "" {
		errValidation["coverImage"] = "Gambar sampul wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	updateData := map[string]interface{}{
		"status":     media.MediaStatusPublished,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateGallery(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "gallery.service",
			"gallery_id": id,
		}).WithError(err).Error("failed to publish gallery")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mempublikasikan galeri", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "gallery", id, gallery.toGalleryResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Galeri berhasil dipublikasikan", nil, nil)
}

func (s *service) UpdateArchivedGallery(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}

	gallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	if gallery.Status == media.MediaStatusArchived {
		return pkg.NewResponse(http.StatusOK, "Galeri sudah diarsipkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     media.MediaStatusArchived,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateGallery(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "gallery.service",
			"gallery_id": id,
		}).WithError(err).Error("failed to archive gallery")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengarsipkan galeri", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "gallery", id, gallery.toGalleryResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Galeri berhasil diarsipkan", nil, nil)
}
