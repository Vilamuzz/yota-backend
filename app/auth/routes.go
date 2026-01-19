package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all auth routes to a router group
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/forget-password", handler.ForgetPassword)
		auth.POST("/reset-password", handler.ResetPassword)
	}
}
