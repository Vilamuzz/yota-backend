package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	limiter_redis "github.com/ulule/limiter/v3/drivers/store/redis"
)

type RateLimitMiddleware struct {
	config      config.RateLimitConfig
	redisClient *redis.Client
	store       limiter.Store
}

func NewRateLimitMiddleware(redisClient *redis.Client) *RateLimitMiddleware {
	middleware := &RateLimitMiddleware{
		config:      config.GetRateLimitConfig(),
		redisClient: redisClient,
	}

	// Initialize store once during creation
	if redisClient != nil {
		store, err := limiter_redis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
			Prefix:          "rate_limit",
			CleanUpInterval: time.Hour,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to create Redis store: %v. Falling back to memory store.\n", err)
			middleware.store = memory.NewStore()
		} else {
			middleware.store = store
		}
	} else {
		middleware.store = memory.NewStore()
	}

	return middleware
}

// getStore returns the configured store (Redis or memory)
func (m *RateLimitMiddleware) getStore() limiter.Store {
	return m.store
}

// RateLimit applies rate limiting based on IP address
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	if !m.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  m.config.RequestsPerMinute,
	}

	store := m.getStore()
	instance := limiter.New(store, rate)
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

	store := m.getStore()
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

	store := m.getStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

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

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

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
	return m.RateLimitWithCustomRate(10, 1*time.Minute)
}

// APIRateLimit applies rate limiting for general API endpoints
func (m *RateLimitMiddleware) APIRateLimit() gin.HandlerFunc {
	return m.RateLimitWithCustomRate(100, 1*time.Minute)
}
