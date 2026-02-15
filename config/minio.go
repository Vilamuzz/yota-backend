package config

import (
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectMinIO() *minio.Client {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSLStr := os.Getenv("MINIO_USE_SSL")

	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		log.Fatal("MinIO configuration (MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY) is not set")
	}

	useSSL := false
	if useSSLStr != "" {
		var err error
		useSSL, err = strconv.ParseBool(useSSLStr)
		if err != nil {
			log.Fatalf("Invalid MINIO_USE_SSL value: %v", err)
		}
	}

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}
