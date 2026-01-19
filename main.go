package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/router"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/docs"
	_ "github.com/Vilamuzz/yota-backend/docs"
	postgre_pkg "github.com/Vilamuzz/yota-backend/pkg/postgre"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	_ = godotenv.Load()
}

func main() {
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

	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	timeoutContext := time.Duration(timeout) * time.Second

	// logger
	writers := make([]io.Writer, 0)
	if logSTDOUT, _ := strconv.ParseBool(os.Getenv("LOG_TO_STDOUT")); logSTDOUT {
		writers = append(writers, os.Stdout)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(io.MultiWriter(writers...))

	// set gin writer to logrus
	gin.DefaultWriter = logrus.StandardLogger().Writer()

	postgre := config.ConnectDB()
	postgre_pkg.AutoMigrateDB(postgre, postgre_pkg.GetAllModels()...)

	// Initialize repositories
	userRepo := user.NewPostgreRepository(postgre)
	authRepo := auth.NewRepository(postgre)

	// Initialize services
	authService := auth.NewService(userRepo, authRepo, timeoutContext)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)

	// Initialize middleware
	appMiddleware := middleware.NewAppMiddleware()

	if os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	ginEngine := gin.New()

	// Apply middleware
	ginEngine.Use(appMiddleware.RecoveryHandler())
	ginEngine.Use(appMiddleware.LoggerHandler(io.MultiWriter(writers...)))

	ginEngine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Ticket-Token"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	// Initialize router
	appRouter := router.NewRouter(authHandler, appMiddleware)
	appRouter.SetupRoutes(ginEngine)

	// Basic routes
	ginEngine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"message": "Welcome to Yota Backend",
		})
	})
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	logrus.Infof("Service running on port %s", port)
	ginEngine.Run(":" + port)
}
