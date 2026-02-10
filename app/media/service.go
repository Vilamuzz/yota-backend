package media

import (
	"context"
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg/minio"
)

type Service interface {
	UploadMedia(ctx context.Context, files []*multipart.FileHeader, prefix string) ([]MediaItem, error)
	DeleteEntityMedia(ctx context.Context, entityID string) error
}

type service struct {
	repo        Repository
	minioClient minio.Client
}

func NewService(repo Repository, minioClient minio.Client) Service {
	return &service{
		repo:        repo,
		minioClient: minioClient,
	}
}

func (s *service) UploadMedia(ctx context.Context, files []*multipart.FileHeader, prefix string) ([]MediaItem, error) {
	var mediaItems []MediaItem

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

		fileURL, err := s.minioClient.UploadFile(ctx, file, folder)
		if err != nil {
			return nil, err
		}

		mediaItems = append(mediaItems, MediaItem{
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
		objectName := minio.ExtractObjectNameFromURL(m.URL)
		if objectName != "" {
			// Delete file from MinIO (ignore error if file doesn't exist)
			_ = s.minioClient.DeleteFile(ctx, objectName)
		}
	}

	// Delete media records from database
	return s.repo.DeleteEntityMedia(ctx, entityID)
}
