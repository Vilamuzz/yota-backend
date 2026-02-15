package pkg

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/gosimple/slug"
)

var allowedImageMimeTypes = []string{
	"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
}

func ImageUploadValidation(file multipart.File, header *multipart.FileHeader) error {
	// get max size config
	maxSize := config.GetMaxFileUploadSize()

	// check max size
	if header.Size > maxSize {
		return fmt.Errorf("file size can not exceed %d mb", maxSize/(1024*1024))
	}

	// get mimetype
	mimeType := header.Header.Get("Content-Type")

	// check mimetype
	isValidType := false
	for _, allowedMimeType := range allowedImageMimeTypes {
		if mimeType == allowedMimeType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid image extension, expected extension: JPG, JPEG, GIF, PNG, WEBP, SVG")
	}

	return nil
}

func GenerateCleanName(originalName string) string {
	nano := time.Now().UnixNano()
	extName := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, extName)
	slugBaseName := slug.Make(baseName)
	name := strconv.Itoa(int(nano)) + "-" + slugBaseName + extName
	return name
}
