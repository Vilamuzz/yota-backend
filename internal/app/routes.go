package app

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (a *App) setupRoutes(engine *gin.Engine) {
	// Apply global middleware
	engine.Use(a.container.Middleware.CORSHandler())
	engine.Use(a.container.Middleware.LoggerHandler(gin.DefaultWriter))
	engine.Use(a.container.Middleware.RecoveryHandler())

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	// Swagger documentation
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes - just register all handlers at once
	api := engine.Group("/api")
	a.container.RegisterHandlers(api)
}
