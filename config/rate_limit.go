package config

import (
	"os"
	"strconv"
)

type RateLimitConfig struct {
	RequestsPerMinute int64
	Enabled           bool
}

func GetRateLimitConfig() RateLimitConfig {
	requestsPerMinute, _ := strconv.ParseInt(os.Getenv("RATE_LIMIT_PER_MINUTE"), 10, 64)
	if requestsPerMinute == 0 {
		requestsPerMinute = 60 // Default: 60 requests per minute
	}

	enabled := os.Getenv("RATE_LIMIT_ENABLED")
	if enabled == "" {
		enabled = "true"
	}

	return RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		Enabled:           enabled == "true",
	}
}
