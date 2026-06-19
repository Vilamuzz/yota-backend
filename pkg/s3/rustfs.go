package s3_pkg

import "strings"

func ExtractObjectNameFromURL(fileURL string) string {
	if !strings.HasPrefix(fileURL, "http://") && !strings.HasPrefix(fileURL, "https://") {
		return fileURL
	}

	// URL format: http://minio:9000/bucket-name/galleries/images/uuid.jpg
	parts := strings.Split(fileURL, "/")
	if len(parts) < 2 {
		return ""
	}

	// Find bucket name and get everything after it
	for i, part := range parts {
		if part != "" && i > 2 { // Skip protocol and domain
			// Join remaining parts as object name
			return strings.Join(parts[i+1:], "/")
		}
	}
	return ""
}
