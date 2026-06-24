package s3_pkg

import (
	"fmt"
	"os"
	"strings"
)

var CDNBaseURL string

// InitCDN must be called explicitly from main() after env vars are loaded
// (e.g. after godotenv.Load()). Using a package-level init() is not safe here
// because Go runs package init()s before main(), so .env would not be loaded yet.
func InitCDN() {
	CDNBaseURL = os.Getenv("CDN_BASE_URL")
	if CDNBaseURL == "" {
		endpoint := os.Getenv("S3_ENDPOINT")
		bucketName := os.Getenv("S3_BUCKET_NAME")
		useSSL := os.Getenv("S3_USE_SSL")
		if endpoint != "" && bucketName != "" {
			protocol := "http"
			if useSSL == "true" {
				protocol = "https"
			}
			CDNBaseURL = fmt.Sprintf("%s://%s/%s", protocol, endpoint, bucketName)
		}
	}
}

// GetCDNURL prepends the CDNBaseURL to the path if it is a relative path.
func GetCDNURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if CDNBaseURL == "" {
		return path
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(CDNBaseURL, "/"), strings.TrimPrefix(path, "/"))
}
