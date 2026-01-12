package middleware

import (
	"io"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/gin-gonic/gin"
)

type appMiddleware struct {
	secretKeySuperadmin string
	secretKeyAdmin      string
	secretKeyUser       string
}

func NewAppMiddleware() AppMiddleware {
	return &appMiddleware{
		secretKeySuperadmin: config.GetJWTSecretKeySuperadmin(),
		secretKeyAdmin:      config.GetJWTSecretKeyAdmin(),
		secretKeyUser:       config.GetJWTSecretKeyUser(),
	}
}

type AppMiddleware interface {
	AuthSuperadmin() gin.HandlerFunc
	AuthAdmin() gin.HandlerFunc
	AuthUser() gin.HandlerFunc
	Logger(writer io.Writer) gin.HandlerFunc
	Recovery() gin.HandlerFunc
}
