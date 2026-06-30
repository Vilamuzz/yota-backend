package pkg

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	ugcPolicy    *bluemonday.Policy
	strictPolicy *bluemonday.Policy
	once         sync.Once
)

// InitPolicies initializes the sanitization policies once.
func InitPolicies() {
	once.Do(func() {
		ugcPolicy = bluemonday.UGCPolicy()
		strictPolicy = bluemonday.StrictPolicy()
	})
}

// SanitizeHTML filters HTML input and leaves only a safe subset of tags (UGC policy).
func SanitizeHTML(input string) string {
	InitPolicies()
	return ugcPolicy.Sanitize(input)
}

// SanitizeStrict strips all HTML elements from the input, returning plain text.
func SanitizeStrict(input string) string {
	InitPolicies()
	return strictPolicy.Sanitize(input)
}
