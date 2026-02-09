package news

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type Service interface {
	FetchAllNews(ctx context.Context, queryParams NewsQueryParams) pkg.Response
	FetchNewsByID(ctx context.Context, id string, incrementView bool) pkg.Response
	CreateNews(ctx context.Context, req NewsRequest) pkg.Response
	UpdateNews(ctx context.Context, id string, req UpdateNewsRequest) pkg.Response
	DeleteNews(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *service) FetchAllNews(ctx context.Context, queryParams NewsQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Set default limit
	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.Category != "" {
		options["category"] = queryParams.Category
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.Cursor != "" {
		options["cursor"] = queryParams.Cursor
	}

	newsList, err := s.repo.FetchAllNews(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch news", nil, nil)
	}

	// Check if there are more results
	hasNext := len(newsList) > queryParams.Limit
	if hasNext {
		newsList = newsList[:queryParams.Limit]
	}

	// Generate cursors
	var nextCursor, prevCursor string
	hasPrev := queryParams.Cursor != ""

	if hasNext && len(newsList) > 0 {
		lastNews := newsList[len(newsList)-1]
		nextCursor = pkg.EncodeCursor(lastNews.CreatedAt, lastNews.ID)
	}

	if hasPrev && len(newsList) > 0 {
		firstNews := newsList[0]
		prevCursor = pkg.EncodeCursor(firstNews.CreatedAt, firstNews.ID)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"news": newsList,
		"pagination": map[string]interface{}{
			"next_cursor": nextCursor,
			"prev_cursor": prevCursor,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"limit":       queryParams.Limit,
		},
	})
}

func (s *service) FetchNewsByID(ctx context.Context, id string, incrementView bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid news ID format"}, nil)
	}

	news, err := s.repo.FetchNewsByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "News not found", nil, nil)
	}

	// Increment views if requested (for public access)
	if incrementView {
		go s.repo.IncrementViews(context.Background(), id)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, news)
}

func (s *service) CreateNews(ctx context.Context, req NewsRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validation
	errValidation := make(map[string]string)

	// Validate title
	if req.Title == "" {
		errValidation["title"] = "Title is required"
	} else if len(req.Title) < 5 {
		errValidation["title"] = "Title must be at least 5 characters"
	} else if len(req.Title) > 200 {
		errValidation["title"] = "Title must not exceed 200 characters"
	}

	// Validate content
	if req.Content == "" {
		errValidation["content"] = "Content is required"
	} else if len(req.Content) < 50 {
		errValidation["content"] = "Content must be at least 50 characters"
	}

	// Validate category
	if req.Category == "" {
		errValidation["category"] = "Category is required"
	} else if req.Category != CategoryGeneral && req.Category != CategoryEvent &&
		req.Category != CategoryAnnouncement && req.Category != CategoryDonation &&
		req.Category != CategorySocial {
		errValidation["category"] = "Invalid category. Must be: general, event, announcement, donation, or social"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = StatusDraft
	}

	// Create news
	timeNow := time.Now()

	var mediaItems []media.Media
	if len(req.Media) > 0 {
		for _, m := range req.Media {
			mediaItems = append(mediaItems, media.Media{
				ID:      uuid.New().String(),
				Type:    m.Type,
				URL:     m.URL,
				AltText: m.AltText,
			})
		}
	}

	news := &News{
		ID:        uuid.New().String(),
		Title:     req.Title,
		Category:  req.Category,
		Content:   req.Content,
		Image:     req.Image,
		Status:    status,
		Views:     0,
		Media:     mediaItems,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}

	// Set published_at if status is published
	if status == StatusPublished {
		news.PublishedAt = &timeNow
	}

	if err := s.repo.CreateOneNews(ctx, news); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create news", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "News successfully created", nil, news)
}

func (s *service) UpdateNews(ctx context.Context, id string, req UpdateNewsRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid news ID format"}, nil)
	}

	// Check if news exists
	existingNews, err := s.repo.FetchNewsByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "News not found", nil, nil)
	}

	// Validation
	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	// Validate and set title
	if req.Title != "" {
		if len(req.Title) < 5 {
			errValidation["title"] = "Title must be at least 5 characters"
		} else if len(req.Title) > 200 {
			errValidation["title"] = "Title must not exceed 200 characters"
		} else {
			updateData["title"] = req.Title
		}
	}

	// Validate and set content
	if req.Content != "" {
		if len(req.Content) < 50 {
			errValidation["content"] = "Content must be at least 50 characters"
		} else {
			updateData["content"] = req.Content
		}
	}

	// Validate and set category
	if req.Category != "" {
		if req.Category != CategoryGeneral && req.Category != CategoryEvent &&
			req.Category != CategoryAnnouncement && req.Category != CategoryDonation &&
			req.Category != CategorySocial {
			errValidation["category"] = "Invalid category. Must be: general, event, announcement, donation, or social"
		} else {
			updateData["category"] = req.Category
		}
	}

	// Validate and set status
	if req.Status != "" {
		if req.Status != StatusDraft && req.Status != StatusPublished && req.Status != StatusArchived {
			errValidation["status"] = "Invalid status. Must be: draft, published, or archived"
		} else {
			updateData["status"] = req.Status

			// Set published_at if status is changing to published and wasn't published before
			if req.Status == StatusPublished && existingNews.PublishedAt == nil {
				now := time.Now()
				updateData["published_at"] = &now
			}
		}
	}

	// Validate and set image
	if req.Image != "" {
		updateData["image"] = req.Image
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	// Set updated_at
	updateData["updated_at"] = time.Now()

	if err := s.repo.UpdateNews(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update news", nil, nil)
	}

	// Update media if provided
	if req.Media != nil {
		var mediaItems []media.Media
		for _, m := range req.Media {
			mediaItems = append(mediaItems, media.Media{
				ID:      uuid.New().String(),
				Type:    m.Type,
				URL:     m.URL,
				AltText: m.AltText,
			})
		}

		newsForKey := &News{ID: id}
		if err := s.repo.UpdateNewsMedia(ctx, newsForKey, mediaItems); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to update news media", nil, nil)
		}
	}

	return pkg.NewResponse(http.StatusOK, "News updated successfully", nil, nil)
}

func (s *service) DeleteNews(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid news ID format"}, nil)
	}

	// Check if news exists
	_, err := s.repo.FetchNewsByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "News not found", nil, nil)
	}

	if err := s.repo.DeleteNews(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete news", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "News deleted successfully", nil, nil)
}
