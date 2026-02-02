package middleware

import (
	"io"

	"github.com/gin-gonic/gin"
)

type AppMiddleware struct {
	Logger   *LoggerMiddleware
	Recovery *RecoveryMiddleware
	JWT      *JWTMiddleware
	CORS     *CORSMiddleware
}

func NewAppMiddleware() *AppMiddleware {
	return &AppMiddleware{
		Logger:   NewLoggerMiddleware(),
		Recovery: NewRecoveryMiddleware(),
		JWT:      NewJWTMiddleware(),
		CORS:     NewCORSMiddleware(),
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
