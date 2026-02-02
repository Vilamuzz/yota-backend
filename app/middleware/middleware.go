package middleware

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AppMiddleware struct {
	Logger    *LoggerMiddleware
	Recovery  *RecoveryMiddleware
	JWT       *JWTMiddleware
	CORS      *CORSMiddleware
	RateLimit *RateLimitMiddleware
}

func NewAppMiddleware(redisClient *redis.Client) *AppMiddleware {
	return &AppMiddleware{
		Logger:    NewLoggerMiddleware(),
		Recovery:  NewRecoveryMiddleware(),
		JWT:       NewJWTMiddleware(),
		CORS:      NewCORSMiddleware(),
		RateLimit: NewRateLimitMiddleware(redisClient),
	}
}

// Convenience methods for easier access
func (m *AppMiddleware) LoggerHandler(writer io.Writer) gin.HandlerFunc {
	return m.Logger.Logger(writer)
}

func (m *AppMiddleware) RecoveryHandler() gin.HandlerFunc {
	return m.Recovery.Recovery()
}

func (m *AppMiddleware) AuthRequired() gin.HandlerFunc {
	return m.JWT.AuthRequired()
}

func (m *AppMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return m.JWT.RequireRoles(roles...)
}

func (m *AppMiddleware) CORSHandler() gin.HandlerFunc {
	return m.CORS.CORS()
}

func (m *AppMiddleware) RateLimitHandler() gin.HandlerFunc {
	return m.RateLimit.StrictRateLimit()
}

func (m *AppMiddleware) AuthRateLimitHandler() gin.HandlerFunc {
	return m.RateLimit.AuthRateLimit()
}

func (m *AppMiddleware) APIRateLimitHandler() gin.HandlerFunc {
	return m.RateLimit.APIRateLimit()
}

func (m *AppMiddleware) CustomRateLimitHandler(requests int64, period time.Duration) gin.HandlerFunc {
	return m.RateLimit.RateLimitWithCustomRate(requests, period)
}
