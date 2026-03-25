package middleware

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RecoveryMiddleware struct{}

func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

func (m *RecoveryMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "recovery.middleware",
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"client_ip": c.ClientIP(),
				}).Error("panic recovered: ", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, pkg.NewResponse(
					http.StatusInternalServerError,
					"Something went wrong",
					nil,
					nil,
				))
			}
		}()
		c.Next()
	}
}
