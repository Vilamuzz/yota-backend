package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimitMiddleware struct {
	config config.RateLimitConfig
}

func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		config: config.GetRateLimitConfig(),
	}
}

// RateLimit applies rate limiting based on IP address
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	if !m.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Define rate: requests per minute
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  m.config.RequestsPerMinute,
	}

	// Create a new memory store
	store := memory.NewStore()

	// Create limiter instance
	instance := limiter.New(store, rate)

	// Create middleware
	middleware := mgin.NewMiddleware(instance)

	return func(c *gin.Context) {
		middleware(c)
	}
}

// RateLimitWithCustomRate creates a rate limiter with custom limits
func (m *RateLimitMiddleware) RateLimitWithCustomRate(requests int64, period time.Duration) gin.HandlerFunc {
	if !m.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate := limiter.Rate{
		Period: period,
		Limit:  requests,
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)
	middleware := mgin.NewMiddleware(instance)

	return func(c *gin.Context) {
		middleware(c)
	}
}

// StrictRateLimit applies strict rate limiting with custom error handling
func (m *RateLimitMiddleware) StrictRateLimit() gin.HandlerFunc {
	if !m.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  m.config.RequestsPerMinute,
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get current context from limiter
		context, err := instance.Get(c, clientIP)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, pkg.NewResponse(
				http.StatusInternalServerError,
				"Rate limiter error",
				nil,
				nil,
			))
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// Check if limit is reached
		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, pkg.NewResponse(
				http.StatusTooManyRequests,
				fmt.Sprintf("Rate limit exceeded. Try again in %v", time.Until(time.Unix(context.Reset, 0))),
				nil,
				map[string]interface{}{
					"limit":     context.Limit,
					"remaining": context.Remaining,
					"reset_at":  time.Unix(context.Reset, 0).Format(time.RFC3339),
				},
			))
			return
		}

		c.Next()
	}
}

// AuthRateLimit applies rate limiting for authentication endpoints (stricter)
func (m *RateLimitMiddleware) AuthRateLimit() gin.HandlerFunc {
	return m.RateLimitWithCustomRate(10, 1*time.Minute) // 10 requests per minute for auth
}

// APIRateLimit applies rate limiting for general API endpoints
func (m *RateLimitMiddleware) APIRateLimit() gin.HandlerFunc {
	return m.RateLimitWithCustomRate(100, 1*time.Minute) // 100 requests per minute for API
}
