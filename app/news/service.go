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
	CreateNews(ctx context.Context, payload NewsCreateRequest) pkg.Response
	UpdateNews(ctx context.Context, id string, payload NewsUpdateRequest) pkg.Response
	DeleteNews(ctx context.Context, id string) pkg.Response
	UpdatePublishNews(ctx context.Context, id string) pkg.Response
	UpdateArchiveNews(ctx context.Context, id string) pkg.Response
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

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(newsList) > params.Limit
		hasNext = true
		if len(newsList) > params.Limit {
			newsList = newsList[:params.Limit]
		}
		// Reverse the slice because the repository returns ASC order for PrevCursor
		for i, j := 0, len(newsList)-1; i < j; i, j = i+1, j-1 {
			newsList[i], newsList[j] = newsList[j], newsList[i]
		}
	} else {
		hasNext = len(newsList) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			newsList = newsList[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toNewsListResponse(newsList, pagination))
}

func (s *service) GetNewsBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"slug": slug, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	go s.repo.IncrementViews(context.Background(), news.ID.String())

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, news.toNewsResponse())
}

func (s *service) GetNewsByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}

	news, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, news.toNewsResponse())
}

func (s *service) CreateNews(ctx context.Context, payload NewsCreateRequest) pkg.Response {
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

	existing, _ := s.repo.FindOneNews(ctx, map[string]interface{}{"title": payload.Title})
	if existing != nil {
		errValidation["title"] = "Berita dengan judul ini sudah ada"
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
	for i, file := range payload.MediaFiles {
		var finalURL string
		if file != nil {
			url, err := s.s3Client.UploadFile(ctx, file, "news")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "news.service",
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

	s.logService.CreateLog(ctx, nil, "CREATE", "news", news.ID.String(), nil, news.toNewsResponse())

	return pkg.NewResponse(http.StatusCreated, "Berita berhasil dibuat", nil, news.toNewsResponse())
}

func (s *service) UpdateNews(ctx context.Context, id string, payload NewsUpdateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
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
		if payload.Title != existingNews.Title {
			existing, _ := s.repo.FindOneNews(ctx, map[string]interface{}{"title": payload.Title})
			if existing != nil {
				errValidation["title"] = "Berita dengan judul ini sudah ada"
			} else {
				finalTitle = payload.Title
				updateData["title"] = payload.Title
				updateData["slug"] = pkg.Slugify(payload.Title)
			}
		}
	}

	finalContent := existingNews.Content
	if payload.Content != "" {
		finalContent = payload.Content
		updateData["content"] = payload.Content
	}

	finalCategory := existingNews.Category
	if payload.Category != "" {
		finalCategory = payload.Category
		updateData["category"] = payload.Category
	}

	if len(finalTitle) < 5 {
		errValidation["title"] = "Judul minimal 5 karakter"
	} else if len(finalTitle) > 200 {
		errValidation["title"] = "Judul maksimal 200 karakter"
	}

	if targetStatus == media.MediaStatusPublished {
		if finalContent == "" {
			errValidation["content"] = "Konten wajib diisi"
		} else if len(finalContent) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		}

		if finalCategory == "" {
			errValidation["category"] = "Kategori wajib diisi"
		}

		if len(payload.MediaFiles) == 0 && len(existingNews.Media) == 0 {
			errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
		}

		if payload.CoverImage == nil && existingNews.CoverImage == "" {
			errValidation["coverImage"] = "Gambar sampul wajib diisi"
		}

		if targetStatus == media.MediaStatusPublished && existingNews.PublishedAt == nil {
			now := time.Now()
			updateData["published_at"] = &now
		}
	} else {
		if finalContent != "" && len(finalContent) < 50 {
			errValidation["content"] = "Konten minimal 50 karakter"
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// Handle Media Updates (Deletions, Updates, and Additions)
	existingMediaList, err := s.mediaService.FetchEntityMedia(ctx, id, "news")
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
					"component": "news.service",
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
					"component": "news.service",
					"media_id":  mid,
				}).WithError(err).Warn("failed to update media metadata")
			}
		}
	}

	// 3. Add New Media Files
	for i, file := range payload.MediaFiles {
		if file != nil {
			url, err := s.s3Client.UploadFile(ctx, file, "news")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "news.service",
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

			if err := s.mediaService.CreateEntityMedia(ctx, id, "news", newMedia); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "news.service",
					"id":        id,
				}).WithError(err).Error("failed to save new media data")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan data media baru", nil, nil)
			}
		}
	}

	if payload.CoverImage != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.CoverImage, "news")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "news.service",
				"news_id":   id,
			}).WithError(err).Error("failed to upload new cover image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar sampul baru", nil, nil)
		}

		if existingNews.CoverImage != "" {
			existingCoverImage := s3_pkg.ExtractObjectNameFromURL(existingNews.CoverImage)
			_ = s.s3Client.DeleteFile(ctx, existingCoverImage)
		}

		updateData["cover_image"] = uploadedURL
	}

	if len(updateData) == 0 && len(payload.MediaFiles) == 0 && payload.CoverImage == nil {
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

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, existingNews.toNewsResponse(), updateData)

	return pkg.NewResponse(http.StatusOK, "Berita berhasil diperbarui", nil, nil)
}

func (s *service) DeleteNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID berita tidak valid"}, nil)
	}
	existingNews, err := s.repo.FindOneNews(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	if err := s.mediaService.DeleteEntityMedia(ctx, id, "news"); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"news_id":   id,
		}).WithError(err).Warn("failed to delete associated media")
	}

	if existingNews.CoverImage != "" {
		objectName := s3_pkg.ExtractObjectNameFromURL(existingNews.CoverImage)
		if objectName != "" {
			_ = s.s3Client.DeleteFile(ctx, objectName)
		}
	}

	if err := s.repo.DeleteNews(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news.service",
			"news_id":   id,
		}).WithError(err).Error("failed to delete news record")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus berita", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "news", id, existingNews.toNewsResponse(), nil)

	return pkg.NewResponse(http.StatusOK, "Berita berhasil dihapus", nil, nil)
}

func (s *service) UpdatePublishNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
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

	if len(news.Media) == 0 {
		errValidation["media"] = "Minimal satu file media wajib diisi untuk publikasi"
	}

	if news.CoverImage == "" {
		errValidation["coverImage"] = "Gambar sampul wajib diisi"
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

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, news.toNewsResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Berita berhasil dipublikasikan", nil, nil)
}

func (s *service) UpdateArchiveNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
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

	s.logService.CreateLog(ctx, nil, "UPDATE", "news", id, news.toNewsResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Berita berhasil diarsipkan", nil, nil)
}
