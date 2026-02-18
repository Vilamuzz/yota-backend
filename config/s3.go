package config

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ConnectS3() *s3.Client {
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

	// Construct endpoint URL with protocol
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	endpointURL := protocol + "://" + endpoint

	// Create custom resolver for S3-compatible endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpointURL,
			SigningRegion:     region,
			HostnameImmutable: true,
		}, nil
	})

	// Create AWS config with custom credentials and endpoint
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS SDK config: %v", err)
	}

	// Create S3 client with path-style addressing (required for MinIO/RustFS)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return s3Client
}
