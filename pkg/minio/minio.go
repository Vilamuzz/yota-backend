package minio

import "strings"

func ExtractObjectNameFromURL(fileURL string) string {
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
