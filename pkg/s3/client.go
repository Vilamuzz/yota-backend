package s3_pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type Client interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	GetFileLink(ctx context.Context, objectName string) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
}

type client struct {
	minioClient *minio.Client
	bucketName  string
	endpoint    string
}

func NewClient(minioClient *minio.Client) Client {
	bucketName := os.Getenv("RUSTFS_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "default-bucket"
	}

	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	useSSL := os.Getenv("RUSTFS_USE_SSL")
	protocol := "http"
	if useSSL == "true" {
		protocol = "https"
	}

	ctx := context.Background()
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
		} else {
			fmt.Printf("Failed to create or verify bucket %s: %v\n", bucketName, err)
		}
	}

	// Set bucket policy to public read
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucketName)

	err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		fmt.Printf("Failed to set bucket policy: %v\n", err)
	}

	return &client{
		minioClient: minioClient,
		bucketName:  bucketName,
		endpoint:    protocol + "://" + endpoint,
	}
}

func (c *client) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Read file content into memory
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// Generate a unique file name
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Prepare upload input
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload the file using bytes.NewReader
	_, err = c.minioClient.PutObject(ctx, c.bucketName, filename, bytes.NewReader(fileContent), int64(len(fileContent)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Construct the public URL
	fileURL := fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucketName, filename)

	return fileURL, nil
}

func (c *client) GetFileLink(ctx context.Context, objectName string) (string, error) {
	// For now, return the direct URL since we're using public bucket
	fileURL := fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucketName, objectName)
	return fileURL, nil
}

func (c *client) DeleteFile(ctx context.Context, objectName string) error {
	err := c.minioClient.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	return err
}
