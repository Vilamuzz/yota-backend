package app

import (
	"os"

	"github.com/Vilamuzz/yota-backend/docs"
	"github.com/Vilamuzz/yota-backend/internal/container"
	"github.com/Vilamuzz/yota-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type App struct {
	engine    *gin.Engine
	container *container.Container
}

func NewApp() (*App, func(), error) {
	// Setup logger
	setupLogger()

	// Setup Swagger
	setupSwagger()

	// Initialize container
	cnt, cleanup, err := container.NewContainer()
	if err != nil {
		return nil, nil, err
	}

	// Setup Gin
	engine := setupGinEngine()

	app := &App{
		engine:    engine,
		container: cnt,
	}

	// Setup routes
	app.setupRoutes(engine)

	// Start background scheduler
	cnt.Scheduler.Start()

	return app, cleanup, nil
}

func (a *App) Run(addr string) error {
	logrus.Infof("Service running on %s", addr)
	return a.engine.Run(addr)
}

func setupLogger() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "yota-backend"
	}
	logger.Setup(appName)
	gin.DefaultWriter = logrus.StandardLogger().Writer()
}

func setupSwagger() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Yota Backend"
	}

	scheme := "http"
	if os.Getenv("APP_ENV") == "production" {
		scheme = "https"
	}

	docs.SwaggerInfo.Title = appName
	docs.SwaggerInfo.Description = "API Documentations"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{scheme}
}

func setupGinEngine() *gin.Engine {
	if os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}
