package config

import (
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectS3() *minio.Client {
	// Read RustFS (S3-compatible) configuration
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	useSSLStr := os.Getenv("S3_USE_SSL")
	region := os.Getenv("S3_REGION")

	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		log.Fatal("RustFS configuration (S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY) is not set")
	}

	// Default region if not specified
	if region == "" {
		region = "us-east-1"
	}

	useSSL := false
	if useSSLStr != "" {
		var err error
		useSSL, err = strconv.ParseBool(useSSLStr)
		if err != nil {
			log.Fatalf("Invalid S3_USE_SSL value: %v", err)
		}
	}

	// Initialize minio client object.
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		log.Fatalln("Failed to load MinIO SDK config:", err)
	}

	return s3Client
}
