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
	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	accessKeyID := os.Getenv("RUSTFS_ACCESS_KEY")
	secretAccessKey := os.Getenv("RUSTFS_SECRET_KEY")
	useSSLStr := os.Getenv("RUSTFS_USE_SSL")
	region := os.Getenv("RUSTFS_REGION")

	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		log.Fatal("RustFS configuration (RUSTFS_ENDPOINT, RUSTFS_ACCESS_KEY, RUSTFS_SECRET_KEY) is not set")
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
			log.Fatalf("Invalid RUSTFS_USE_SSL value: %v", err)
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
