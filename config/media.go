package config

import (
	"os"
	"strconv"
)

func GetMaxFileUploadSize() int64 {
	maxFileUploadSize, _ := strconv.ParseInt(os.Getenv("MAX_FILE_SIZE"), 10, 64)
	if maxFileUploadSize <= 0 {
		maxFileUploadSize = 5 // default 5 mb
	}
	return maxFileUploadSize * 1024 * 1024
}
