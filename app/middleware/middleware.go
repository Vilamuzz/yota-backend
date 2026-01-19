package middleware

import (
	"io"

	"github.com/gin-gonic/gin"
)

type AppMiddleware struct {
	Logger   *LoggerMiddleware
	Recovery *RecoveryMiddleware
	JWT      *JWTMiddleware
}

func NewAppMiddleware() *AppMiddleware {
	return &AppMiddleware{
		Logger:   NewLoggerMiddleware(),
		Recovery: NewRecoveryMiddleware(),
		JWT:      NewJWTMiddleware(),
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
