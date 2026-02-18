package media

import (
	"context"
	"mime/multipart"

	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type Service interface {
	UploadMedia(ctx context.Context, files []*multipart.FileHeader, prefix string) ([]MediaRequest, error)
	DeleteEntityMedia(ctx context.Context, entityID string) error
	CreateEntityMedia(ctx context.Context, entityID, entityType string, mediaRequests []MediaRequest) error
	DeleteMediaByID(ctx context.Context, mediaID string) error
	FetchEntityMedia(ctx context.Context, entityID, entityType string) ([]Media, error)
	UpdateMediaByID(ctx context.Context, mediaID string, updateData map[string]interface{}) error
}

type service struct {
	repo        Repository
	s3Client     s3_pkg.Client
}

func NewService(repo Repository, s3Client s3_pkg.Client) Service {
	return &service{
		repo:        repo,
		s3Client:     s3Client,
	}
}

func (s *service) UploadMedia(ctx context.Context, files []*multipart.FileHeader, prefix string) ([]MediaRequest, error) {
	var mediaItems []MediaRequest

	for _, file := range files {
		// Determine folder based on file type (image or video)
		folder := prefix + "/others"
		mediaType := MediaTypeImage // Default
		mimeType := file.Header.Get("Content-Type")

		if len(mimeType) >= 5 && mimeType[:5] == "image" {
			folder = prefix + "/images"
			mediaType = MediaTypeImage
		} else if len(mimeType) >= 5 && mimeType[:5] == "video" {
			folder = prefix + "/videos"
			mediaType = MediaTypeVideo
		}

		fileURL, err := s.s3Client.UploadFile(ctx, file, folder)
		if err != nil {
			return nil, err
		}

		mediaItems = append(mediaItems, MediaRequest{
			URL:     fileURL,
			Type:    mediaType,
			AltText: file.Filename,
		})
	}

	return mediaItems, nil
}

func (s *service) DeleteEntityMedia(ctx context.Context, entityID string) error {
	// Fetch media to get file URLs
	mediaList, err := s.repo.FetchEntityMedia(ctx, entityID)
	if err != nil {
		return err
	}

	// Delete files from MinIO
	for _, m := range mediaList {
		// Extract object name from URL
		// URL format: http://minio:9000/bucket-name/path/to/file.jpg
		objectName := s3_pkg.ExtractObjectNameFromURL(m.URL)
		if objectName != "" {
			// Delete file from MinIO (ignore error if file doesn't exist)
			_ = s.s3Client.DeleteFile(ctx, objectName)
		}
	}

	// Delete media records from database
	return s.repo.DeleteEntityMedia(ctx, entityID)
}

func (s *service) CreateEntityMedia(ctx context.Context, entityID, entityType string, mediaRequests []MediaRequest) error {
	var mediaItems []Media
	for _, m := range mediaRequests {
		mediaItems = append(mediaItems, Media{
			ID:      m.ID,
			Type:    m.Type,
			URL:     m.URL,
			AltText: m.AltText,
		})
	}
	return s.repo.CreateEntityMedia(ctx, entityID, entityType, mediaItems)
}

func (s *service) DeleteMediaByID(ctx context.Context, mediaID string) error {
	// Delete media and get its info for MinIO cleanup
	media, err := s.repo.DeleteMediaByID(ctx, mediaID)
	if err != nil {
		return err
	}

	// Delete file from MinIO
	objectName := s3_pkg.ExtractObjectNameFromURL(media.URL)
	if objectName != "" {
		// Ignore error if file doesn't exist
		_ = s.s3Client.DeleteFile(ctx, objectName)
	}

	return nil
}

func (s *service) FetchEntityMedia(ctx context.Context, entityID, entityType string) ([]Media, error) {
	return s.repo.FetchEntityMedia(ctx, entityID)
}

func (s *service) UpdateMediaByID(ctx context.Context, mediaID string, updateData map[string]interface{}) error {
	return s.repo.UpdateMediaByID(ctx, mediaID, updateData)
}
