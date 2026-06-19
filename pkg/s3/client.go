package s3_pkg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	_ "golang.org/x/image/webp"
)

type Client interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadFileOriginal(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadFileFromBytes(ctx context.Context, fileContent []byte, originalFilename string, contentType string, folder string) (string, error)
	GetFileLink(ctx context.Context, objectName string) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
}

type client struct {
	minioClient *minio.Client
	bucketName  string
	endpoint    string
}

func NewClient(minioClient *minio.Client) Client {
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "default-bucket"
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	useSSL := os.Getenv("S3_USE_SSL")
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
	return c.upload(ctx, file, folder, true)
}

func (c *client) UploadFileOriginal(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	return c.upload(ctx, file, folder, false)
}

func (c *client) UploadFileFromBytes(ctx context.Context, fileContent []byte, originalFilename string, contentType string, folder string) (string, error) {
	return c.uploadBytes(ctx, fileContent, originalFilename, contentType, folder, true)
}

func (c *client) upload(ctx context.Context, file *multipart.FileHeader, folder string, compress bool) (string, error) {
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

	contentType := file.Header.Get("Content-Type")
	return c.uploadBytes(ctx, fileContent, file.Filename, contentType, folder, compress)
}

func (c *client) uploadBytes(ctx context.Context, fileContent []byte, originalFilename string, contentType string, folder string, compress bool) (string, error) {
	// Generate a unique file name
	ext := strings.ToLower(filepath.Ext(originalFilename))

	// Prepare upload input
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Convert any decodable image format to JPEG at 80% quality.
	isImageUpload := strings.HasPrefix(contentType, "image/") ||
		ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
	if compress && isImageUpload {
		img, _, err := image.Decode(bytes.NewReader(fileContent))
		if err == nil {
			buf := new(bytes.Buffer)
			if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 80}); err == nil {
				fileContent = buf.Bytes()
				ext = ".jpg"
				contentType = "image/jpeg"
			}
		}
	}

	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload the file using bytes.NewReader
	_, err := c.minioClient.PutObject(ctx, c.bucketName, filename, bytes.NewReader(fileContent), int64(len(fileContent)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (c *client) GetFileLink(ctx context.Context, objectName string) (string, error) {
	return GetCDNURL(objectName), nil
}

func (c *client) DeleteFile(ctx context.Context, objectName string) error {
	err := c.minioClient.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	return err
}
