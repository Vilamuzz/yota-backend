package s3_pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type Client interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	GetFileLink(ctx context.Context, objectName string) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
}

type client struct {
	s3Client   *s3.Client
	bucketName string
	endpoint   string
}

func NewClient(s3Client *s3.Client) Client {
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

	// Ensure bucket exists
	ctx := context.Background()
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if bucket already exists
		_, headErr := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if headErr != nil {
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

	_, err = s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	})
	if err != nil {
		fmt.Printf("Failed to set bucket policy: %v\n", err)
	}

	return &client{
		s3Client:   s3Client,
		bucketName: bucketName,
		endpoint:   protocol + "://" + endpoint,
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
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
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

	// If you need presigned URLs, uncomment below:
	/*
		presignClient := s3.NewPresignClient(c.s3Client)
		presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(objectName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = 1000 * time.Second
		})
		if err != nil {
			return "", err
		}
		return presignResult.URL, nil
	*/
}

func (c *client) DeleteFile(ctx context.Context, objectName string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectName),
	})
	return err
}
