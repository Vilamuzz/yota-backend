package news

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
	GetNewsList(ctx context.Context, params NewsQueryParams, isAdmin bool) pkg.Response
	GetNewsBySlug(ctx context.Context, slug string) pkg.Response
	GetNewsByID(ctx context.Context, id string) pkg.Response
	CreateNews(ctx context.Context, payload NewsRequest) pkg.Response
	UpdateNews(ctx context.Context, id string, payload NewsRequest) pkg.Response
	DeleteNews(ctx context.Context, id string) pkg.Response
	UpdatePublishedNews(ctx context.Context, id string) pkg.Response
	UpdateArchivedNews(ctx context.Context, id string) pkg.Response
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

func (s *service) GetNewsList(ctx context.Context, params NewsQueryParams, isAdmin bool) pkg.Response {
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

	newsList, err := s.repo.FindAllNews(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data berita", nil, nil)
	}

	hasNext := len(newsList) > params.Limit
	if hasNext {
		newsList = newsList[:params.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := params.PrevCursor != ""

	if len(newsList) > 0 {
		first := newsList[0]
		last := newsList[len(newsList)-1]
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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, map[string]interface{}{
		"news":       newsList,
		"pagination": pagination,
	})
}

func (s *service) GetNewsBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	go s.repo.IncrementViews(context.Background(), news.ID.String())

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, news)
}

func (s *service) GetNewsByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, news)
}

func (s *service) CreateNews(ctx context.Context, payload NewsRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	status := media.MediaStatusDraft
	if payload.Status != "" {
		status = payload.Status
	}

	if payload.Title == "" {
		errValidation["title"] = "Judul wajib diisi"
	} else if len(payload.Title) < 5 {
		errValidation["title"] = "Judul minimal 5 karakter"
	} else if len(payload.Title) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if status == media.MediaStatusPublished {
		if payload.Content == "" {
			errValidation["content"] = "Konten wajib diisi"
		} else if len(payload.Content) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		}

		if payload.Category == "" {
			errValidation["category"] = "Kategori wajib diisi"
		}

		if payload.CoverImage == nil {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}
	} else {
		if payload.Content != "" && len(payload.Content) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var coverImageURL string
	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "news")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "news.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar sampul", nil, nil)
		}
		coverImageURL = uploadedURL
	}

	var mediaItems []media.Media
	if len(payload.Files) > 0 {
		uploadedMediaItems, err := s.mediaService.UploadMedia(ctx, payload.Files, "news")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "news.service",
				"title":     payload.Title,
			}).WithError(err).Error("failed to upload media")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file media", nil, nil)
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

			mediaItems = append(mediaItems, item)
		}
	}

	timeNow := time.Now()
	news := &News{
		ID:         uuid.New(),
		Title:      payload.Title,
		Slug:       pkg.Slugify(payload.Title),
		Category:   payload.Category,
		Content:    payload.Content,
		CoverImage: coverImageURL,
		Status:     status,
		Media:      mediaItems,
		Views:      0,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
	}

	if status == media.MediaStatusPublished {
		news.PublishedAt = &timeNow
	}

	if err := s.repo.CreateNews(ctx, news); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"title":     payload.Title,
		}).WithError(err).Error("failed to create news")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat berita", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "news", news.ID.String(), nil, news)

	return pkg.NewResponse(http.StatusCreated, "Berita berhasil dibuat", nil, news)
}

func (s *service) UpdateNews(ctx context.Context, id string, payload NewsRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	existingNews, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	targetStatus := existingNews.Status
	if payload.Status != "" {
		targetStatus = payload.Status
		updateData["status"] = payload.Status
	}

	finalTitle := existingNews.Title
	if payload.Title != "" {
		if len(payload.Title) < 5 {
			errValidation["title"] = "Judul minimal 5 karakter"
		} else if len(payload.Title) > 200 {
			errValidation["title"] = "Judul maksimal 200 karakter"
		} else {
			finalTitle = payload.Title
			updateData["title"] = payload.Title
			updateData["slug"] = pkg.Slugify(payload.Title)
		}
	}

	finalContent := existingNews.Content
	if payload.Content != "" {
		if len(payload.Content) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		} else {
			finalContent = payload.Content
			updateData["content"] = payload.Content
		}
	}

	finalCategory := existingNews.Category
	if payload.Category != "" {
		finalCategory = payload.Category
		updateData["category"] = payload.Category
	}

	if targetStatus == media.MediaStatusPublished {
		if finalTitle == "" || len(finalTitle) < 5 {
			errValidation["title"] = "Judul minimal 5 karakter"
		}
		if finalContent == "" {
			errValidation["content"] = "Konten wajib diisi"
		} else if len(finalContent) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		}
		if finalCategory == "" {
			errValidation["category"] = "Kategori wajib diisi"
		}
		if existingNews.CoverImage == "" && payload.CoverImage == nil {
			errValidation["coverImage"] = "Gambar sampul wajib diisi untuk publikasi"
		}
		if payload.Status == media.MediaStatusPublished && existingNews.PublishedAt == nil {
			now := time.Now()
			updateData["published_at"] = &now
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// Handle cover image upload
	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "news")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "news.service",
				"news_id":   id,
			}).WithError(err).Error("failed to upload cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar sampul", nil, nil)
		}

		// Delete old cover image from S3
		if existingNews.CoverImage != "" {
			objectName := s3_pkg.ExtractObjectNameFromURL(existingNews.CoverImage)
			_ = s.s3Client.DeleteFile(ctx, objectName)
		}
		updateData["cover_image"] = uploadedURL
	}

	// Handle multiple media management
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
			var err error
			uploadedMedia, err = s.mediaService.UploadMedia(ctx, payload.Files, "news")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "news.service",
					"news_id":   id,
				}).WithError(err).Error("failed to upload media")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file media", nil, nil)
			}

			for i, item := range uploadedMedia {
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

		existingMediaList, err := s.mediaService.FetchEntityMedia(ctx, id, "news")
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
					"component": "news.service",
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
			if err := s.mediaService.CreateEntityMedia(ctx, id, "news", newMediaItems); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "news.service",
					"news_id":   id,
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

		if err := s.repo.UpdateNews(ctx, id, updateData); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "news.service",
				"news_id":   id,
			}).WithError(err).Error("failed to update news")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui berita", nil, nil)
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, existingNews, updateData)

	return pkg.NewResponse(http.StatusOK, "Berita berhasil diperbarui", nil, nil)
}

func (s *service) DeleteNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	existingNews, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	if err := s.repo.DeleteNews(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"news_id":   id,
		}).WithError(err).Error("failed to delete news")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus berita", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "news", id, existingNews, nil)

	return pkg.NewResponse(http.StatusOK, "Berita berhasil dihapus", nil, nil)
}

func (s *service) UpdatePublishedNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	if news.Status == media.MediaStatusPublished {
		return pkg.NewResponse(http.StatusOK, "Berita sudah dipublikasikan", nil, nil)
	}

	errValidation := make(map[string]string)
	if len(news.Title) < 5 {
		errValidation["title"] = "Judul minimal 5 karakter"
	}
	if news.Content == "" {
		errValidation["content"] = "Konten wajib diisi"
	} else if len(news.Content) < 50 {
		errValidation["content"] = "Konten minimal 50 karakter"
	}
	if news.Category == "" {
		errValidation["category"] = "Kategori wajib diisi"
	}
	if news.CoverImage == "" {
		errValidation["file"] = "Gambar wajib diisi untuk publikasi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	updateData := map[string]interface{}{
		"status":     media.MediaStatusPublished,
		"updated_at": time.Now(),
	}
	if news.PublishedAt == nil {
		now := time.Now()
		updateData["published_at"] = &now
	}

	if err := s.repo.UpdateNews(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"news_id":   id,
		}).WithError(err).Error("failed to publish news")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mempublikasikan berita", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, news, updateData)
	return pkg.NewResponse(http.StatusOK, "Berita berhasil dipublikasikan", nil, nil)
}

func (s *service) UpdateArchivedNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	if news.Status == media.MediaStatusArchived {
		return pkg.NewResponse(http.StatusOK, "Berita sudah diarsipkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     media.MediaStatusArchived,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateNews(ctx, id, updateData); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"news_id":   id,
		}).WithError(err).Error("failed to archive news")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengarsipkan berita", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, news, updateData)
	return pkg.NewResponse(http.StatusOK, "Berita berhasil diarsipkan", nil, nil)
}
