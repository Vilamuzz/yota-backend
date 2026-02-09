package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"

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
}

func NewClient(minioClient *minio.Client) Client {
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default-bucket"
	}

	// Ensure bucket exists
	ctx := context.Background()
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if it exists)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			// Bucket already exists
		} else {
			// Failed to create bucket
			fmt.Printf("Failed to create bucket %s: %v\n", bucketName, err)
		}
	}

	// Set bucket policy to public read (simplified for this use case, adjust for production)
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

	if err := minioClient.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		fmt.Printf("Failed to set bucket policy: %v\n", err)
	}

	return &client{
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

func (c *client) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate a unique file name
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload the file
	_, err = c.minioClient.PutObject(ctx, c.bucketName, filename, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	// Construct the URL
	// Note: In a real production environment, you might want to use a CDN or a specific public URL.
	// For this setup, we'll construct the URL based on the MinIO endpoint.
	endpoint := c.minioClient.EndpointURL()
	fileURL := fmt.Sprintf("%s/%s/%s", endpoint.String(), c.bucketName, filename)

	return fileURL, nil
}

func (c *client) GetFileLink(ctx context.Context, objectName string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.minioClient.PresignedGetObject(ctx, c.bucketName, objectName, time.Duration(1000)*time.Second, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (c *client) DeleteFile(ctx context.Context, objectName string) error {
	return c.minioClient.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
}
