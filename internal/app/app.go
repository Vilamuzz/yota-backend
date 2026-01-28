package app

import (
	"io"
	"os"
	"strconv"

	"github.com/Vilamuzz/yota-backend/docs"
	"github.com/Vilamuzz/yota-backend/internal/container"
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

	return app, cleanup, nil
}

func (a *App) Run(addr string) error {
	logrus.Infof("Service running on %s", addr)
	return a.engine.Run(addr)
}

func setupLogger() {
	writers := make([]io.Writer, 0)
	if logSTDOUT, _ := strconv.ParseBool(os.Getenv("LOG_TO_STDOUT")); logSTDOUT {
		writers = append(writers, os.Stdout)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(io.MultiWriter(writers...))
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
