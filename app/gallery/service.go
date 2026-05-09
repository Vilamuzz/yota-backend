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
	CreateGallery(ctx context.Context, payload GalleryRequest) pkg.Response
	UpdateGallery(ctx context.Context, id string, payload GalleryRequest) pkg.Response
	DeleteGallery(ctx context.Context, id string) pkg.Response
	UpdatePublishedGallery(ctx context.Context, id string) pkg.Response
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

	hasNext := len(galleries) > params.Limit
	if hasNext {
		galleries = galleries[:params.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := params.PrevCursor != ""

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

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}

	gallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, gallery.toGalleryResponse())
}

func (s *service) CreateGallery(ctx context.Context, payload GalleryRequest) pkg.Response {
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
		if len(payload.Files) == 0 {
			errValidation["media"] = "Minimal satu file media wajib diisi"
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

	var mediaItems []media.Media
	var coverImageURL string
	if len(payload.Files) > 0 {
		uploadedMediaItems, err := s.mediaService.UploadMedia(ctx, payload.Files, "galleries")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "gallery.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload media")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file", nil, nil)
		}

		for i, uploadedItem := range uploadedMediaItems {
			item := media.Media{
				ID:      uuid.New(),
				Type:    uploadedItem.Type,
				URL:     uploadedItem.URL,
				AltText: uploadedItem.AltText,
				Order:   i + 1,
			}

			if i < len(payload.Metadata) {
				if payload.Metadata[i].AltText != "" {
					item.AltText = payload.Metadata[i].AltText
				}
				if payload.Metadata[i].Order != 0 {
					item.Order = payload.Metadata[i].Order
				}
			}

			if i == 0 {
				coverImageURL = uploadedItem.URL
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

func (s *service) UpdateGallery(ctx context.Context, id string, payload GalleryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
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
		finalTitle = payload.Title
		updateData["title"] = payload.Title
		updateData["slug"] = pkg.Slugify(payload.Title)
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

		if len(payload.Files) == 0 && len(existingGallery.Media) == 0 {
			if len(payload.Metadata) > 0 {
				hasMedia := false
				for _, m := range payload.Metadata {
					if m.ID != "" {
						hasMedia = true
						break
					}
				}
				if !hasMedia && len(payload.Files) == 0 {
					errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
				}
			} else {
				errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
			}
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

	if len(payload.Metadata) > 0 || len(payload.Files) > 0 {
		var existingMediaMetadata []media.MediaMetadata
		var newMediaMetadata []media.MediaMetadata

		for _, m := range payload.Metadata {
			if m.ID != "" {
				existingMediaMetadata = append(existingMediaMetadata, m)
			} else {
				newMediaMetadata = append(newMediaMetadata, m)
			}
		}

		var uploadedMedia []media.MediaRequest
		if len(payload.Files) > 0 {
			uploadedMedia, err = s.mediaService.UploadMedia(ctx, payload.Files, "galleries")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component":  "gallery.service",
					"gallery_id": id,
				}).WithError(err).Error("failed to upload media")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file", nil, nil)
			}

			for i, item := range uploadedMedia {
				// Set default order for new items
				item.Order = i + 1
				if i < len(newMediaMetadata) {
					if newMediaMetadata[i].AltText != "" {
						item.AltText = newMediaMetadata[i].AltText
					}
					if newMediaMetadata[i].Order != 0 {
						item.Order = newMediaMetadata[i].Order
					}
				}
				uploadedMedia[i] = item
			}
		}

		existingMediaList, err := s.mediaService.FetchEntityMedia(ctx, id, "galleries")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data media yang ada", nil, nil)
		}
		keepMediaIDs := make(map[string]media.MediaMetadata)
		for _, m := range existingMediaMetadata {
			keepMediaIDs[m.ID] = m
		}

		for _, existingMedia := range existingMediaList {
			if _, shouldKeep := keepMediaIDs[existingMedia.ID.String()]; !shouldKeep {
				if err := s.mediaService.DeleteMediaByID(ctx, existingMedia.ID.String()); err != nil {
					continue
				}
			}
		}

		for _, m := range existingMediaMetadata {
			updateMediaData := map[string]interface{}{
				"alt_text": m.AltText,
				"order":    m.Order,
			}
			if err := s.mediaService.UpdateMediaByID(ctx, m.ID, updateMediaData); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "gallery.service",
					"media_id":  m.ID,
				}).WithError(err).Error("failed to update media metadata")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui data media", nil, nil)
			}
		}

		if len(uploadedMedia) > 0 {
			var newMediaItems []media.MediaRequest
			for _, m := range uploadedMedia {
				newMediaItems = append(newMediaItems, media.MediaRequest{
					ID:      uuid.New(),
					URL:     m.URL,
					Type:    m.Type,
					AltText: m.AltText,
					Order:   m.Order,
				})
			}
			if err := s.mediaService.CreateEntityMedia(ctx, id, "galleries", newMediaItems); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":  "gallery.service",
					"gallery_id": id,
				}).WithError(err).Error("failed to create entity media")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat data media baru", nil, nil)
			}
		}
	}

	if len(updateData) == 0 && len(payload.Metadata) == 0 && len(payload.Files) == 0 {
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

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID galeri tidak valid"}, nil)
	}
	existingGallery, err := s.repo.FindOneGallery(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Galeri tidak ditemukan", nil, nil)
	}

	if err := s.repo.DeleteGallery(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "gallery.service",
			"gallery_id": id,
		}).WithError(err).Error("failed to delete gallery")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus galeri", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "gallery", id, existingGallery.toGalleryResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Galeri berhasil dihapus", nil, nil)
}

func (s *service) UpdatePublishedGallery(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
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

	if _, err := uuid.Parse(id); err != nil {
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
