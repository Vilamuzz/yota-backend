package gallery

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

// Note: complete donation transaction pagination, and considering the published gallery using offset based pagination
type Service interface {
	ListPublished(ctx context.Context, queryParams GalleryQueryParams) pkg.Response
	GetPublishedByID(ctx context.Context, id string, incrementView bool) pkg.Response
	List(ctx context.Context, queryParams GalleryQueryParams) pkg.Response
	GetByID(ctx context.Context, id string) pkg.Response
	CreateGallery(ctx context.Context, req GalleryRequest) pkg.Response
	UpdateGallery(ctx context.Context, id string, req UpdateGalleryRequest) pkg.Response
	DeleteGallery(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo         Repository
	timeout      time.Duration
	mediaRepo    media.Repository
	mediaService media.Service
}

func NewService(repo Repository, mediaService media.Service, timeout time.Duration) Service {
	return &service{
		repo:         repo,
		mediaService: mediaService,
		timeout:      timeout,
	}
}

func (s *service) ListPublished(ctx context.Context, queryParams GalleryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	usingPrevCursor := queryParams.PrevCursor != ""

	options := map[string]interface{}{
		"limit":     queryParams.Limit,
		"published": true,
	}
	if queryParams.CategoryID != 0 {
		options["category_id"] = queryParams.CategoryID
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	galleries, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch galleries", nil, nil)
	}

	hasMore := len(galleries) > queryParams.Limit
	if hasMore {
		galleries = galleries[:queryParams.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(galleries)-1; i < j; i, j = i+1, j-1 {
			galleries[i], galleries[j] = galleries[j], galleries[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && queryParams.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && queryParams.NextCursor != "")

	if len(galleries) > 0 {
		first := galleries[0]
		last := galleries[len(galleries)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toPublishedGalleryListResponse(galleries, pagination))
}

func (s *service) GetPublishedByID(ctx context.Context, id string, incrementView bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}

	gallery, err := s.repo.FindByID(ctx, map[string]interface{}{"id": id, "published": true})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
	}

	galleryResponse := gallery.toPublishedGalleryResponse()

	if incrementView {
		go s.repo.IncrementViews(context.Background(), id)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, galleryResponse)
}

func (s *service) List(ctx context.Context, queryParams GalleryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	usingPrevCursor := queryParams.PrevCursor != ""

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.CategoryID != 0 {
		options["category_id"] = queryParams.CategoryID
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	galleries, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch galleries", nil, nil)
	}

	hasMore := len(galleries) > queryParams.Limit
	if hasMore {
		galleries = galleries[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && queryParams.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && queryParams.NextCursor != "")

	if len(galleries) > 0 {
		first := galleries[0]
		last := galleries[len(galleries)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      queryParams.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, toGalleryListResponse(galleries, pagination))
}

func (s *service) GetByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	gallery, err := s.repo.FindByID(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, gallery.toGalleryResponse())
}

func (s *service) CreateGallery(ctx context.Context, req GalleryRequest) pkg.Response {
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
	} else if len(req.Description) > 1000 {
		errValidation["description"] = "Description must not exceed 1000 characters"
	}

	if req.CategoryID == 0 {
		errValidation["category"] = "Category is required"
	}
	if len(req.Files) == 0 {
		errValidation["media"] = "At least one media file is required"
	}

	if len(req.Metadata) > 0 {
		if len(req.Files) != len(req.Metadata) {
			errValidation["metadata"] = "Number of files must match number of metadata entries"
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var mediaItems []media.Media
	if len(req.Files) > 0 {
		uploadedMediaItems, err := s.mediaService.UploadMedia(ctx, req.Files, "galleries")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload files", nil, nil)
		}

		for i, uploadedItem := range uploadedMediaItems {
			item := media.Media{
				ID:      uuid.New().String(),
				Type:    uploadedItem.Type,
				URL:     uploadedItem.URL,
				AltText: "",
				Order:   0,
			}

			if i < len(req.Metadata) {
				item.AltText = req.Metadata[i].AltText
				item.Order = req.Metadata[i].Order
			}

			mediaItems = append(mediaItems, item)
		}
	}

	timeNow := time.Now()

	var publishedAt *time.Time
	if *req.Published {
		publishedAt = &timeNow
	}

	gallery := &Gallery{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Slug:        pkg.Slugify(req.Title),
		CategoryID:  req.CategoryID,
		Description: req.Description,
		PublishedAt: publishedAt,
		Views:       0,
		Media:       mediaItems,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	if err := s.repo.CreateOneGallery(ctx, gallery); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create gallery", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Gallery successfully created", nil, gallery.toGalleryResponse())
}

func (s *service) UpdateGallery(ctx context.Context, id string, req UpdateGalleryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}

	existingGallery, err := s.repo.FindByID(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
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
			updateData["slug"] = pkg.Slugify(req.Title)
		}
	}

	if req.Description != "" {
		if len(req.Description) < 10 {
			errValidation["description"] = "Description must be at least 10 characters"
		} else if len(req.Description) > 1000 {
			errValidation["description"] = "Description must not exceed 1000 characters"
		} else {
			updateData["description"] = req.Description
		}
	}

	if req.CategoryID != 0 {
		updateData["category_id"] = req.CategoryID
	}

	if req.Published != nil {
		if *req.Published {
			if existingGallery.PublishedAt == nil {
				now := time.Now()
				updateData["published_at"] = &now
			}
		} else {
			updateData["published_at"] = nil
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if len(req.Metadata) > 0 {
		var existingMediaMetadata []media.MediaMetadata
		var newMediaMetadata []media.MediaMetadata

		for _, m := range req.Metadata {
			if m.ID != "" {
				existingMediaMetadata = append(existingMediaMetadata, m)
			} else {
				newMediaMetadata = append(newMediaMetadata, m)
			}
		}

		if len(newMediaMetadata) != len(req.Files) {
			errValidation["metadata"] = "Number of new media entries must match number of uploaded files"
			return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
		}
		var uploadedMedia []media.MediaRequest
		if len(req.Files) > 0 {
			uploadedMedia, err = s.mediaService.UploadMedia(ctx, req.Files, "galleries")
			if err != nil {
				return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload files", nil, nil)
			}

			for i, item := range uploadedMedia {
				if i < len(newMediaMetadata) {
					item.AltText = newMediaMetadata[i].AltText
					item.Order = newMediaMetadata[i].Order
					uploadedMedia[i] = item
				}
			}
		}

		existingMediaList, err := s.mediaService.FetchEntityMedia(ctx, id, "galleries")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch existing media", nil, nil)
		}
		keepMediaIDs := make(map[string]media.MediaMetadata)
		for _, m := range existingMediaMetadata {
			keepMediaIDs[m.ID] = m
		}

		for _, existingMedia := range existingMediaList {
			if _, shouldKeep := keepMediaIDs[existingMedia.ID]; !shouldKeep {
				if err := s.mediaService.DeleteMediaByID(ctx, existingMedia.ID); err != nil {
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
				return pkg.NewResponse(http.StatusInternalServerError, "Failed to update media", nil, nil)
			}
		}

		if len(uploadedMedia) > 0 {
			var newMediaItems []media.MediaRequest
			for _, m := range uploadedMedia {
				newMediaItems = append(newMediaItems, media.MediaRequest{
					ID:      uuid.New().String(),
					URL:     m.URL,
					Type:    m.Type,
					AltText: m.AltText,
					Order:   m.Order,
				})
			}
			if err := s.mediaService.CreateEntityMedia(ctx, id, "galleries", newMediaItems); err != nil {
				return pkg.NewResponse(http.StatusInternalServerError, "Failed to create new media", nil, nil)
			}
		}
	}

	if len(updateData) == 0 && len(req.Metadata) == 0 && len(req.Files) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	if len(updateData) > 0 {
		updateData["updated_at"] = time.Now()

		if err := s.repo.UpdateGallery(ctx, id, updateData); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to update gallery", nil, nil)
		}
	}

	return pkg.NewResponse(http.StatusOK, "Gallery updated successfully", nil, nil)
}

func (s *service) DeleteGallery(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid gallery ID format"}, nil)
	}
	_, err := s.repo.FindByID(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Gallery not found", nil, nil)
	}

	if err := s.repo.SoftDeleteGallery(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete gallery", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Gallery deleted successfully", nil, nil)
}
