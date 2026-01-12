package middleware

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (c *appMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Error("Panic Recover : ", err)
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
