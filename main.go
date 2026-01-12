package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	admin_http "github.com/Vilamuzz/yota-backend/app/delivery/http/admin"
	"github.com/Vilamuzz/yota-backend/app/delivery/http/middleware"
	superadmin_http "github.com/Vilamuzz/yota-backend/app/delivery/http/superadmin"
	user_http "github.com/Vilamuzz/yota-backend/app/delivery/http/user"
	postgre_repository "github.com/Vilamuzz/yota-backend/app/repository/postgre"
	admin_usecase "github.com/Vilamuzz/yota-backend/app/usecase/admin"
	superadmin_usecase "github.com/Vilamuzz/yota-backend/app/usecase/superadmin"
	user_usecase "github.com/Vilamuzz/yota-backend/app/usecase/user"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/docs"
	_ "github.com/Vilamuzz/yota-backend/docs"
	"github.com/Vilamuzz/yota-backend/domain"
	"github.com/Vilamuzz/yota-backend/pkg"
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
	docs.SwaggerInfo.Title = appName
	docs.SwaggerInfo.Description = "API Documentations"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

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
	pkg.AutoMigrateDB(postgre, domain.GetAllModels()...)
	postgreDbRepo := postgre_repository.NewPostgreDBRepo(postgre)

	superadminUsecase := superadmin_usecase.NewSuperadminAppUsecase(&superadmin_usecase.RepoInjection{
		PostgreDBRepo: postgreDbRepo,
	}, timeoutContext)

	adminUsecase := admin_usecase.NewAdminAppUsecase(&admin_usecase.RepoInjection{
		PostgreDBRepo: postgreDbRepo,
	}, timeoutContext)

	userUsecase := user_usecase.NewUserAppUsecase(&user_usecase.RepoInjection{
		PostgreDBRepo: postgreDbRepo,
	}, timeoutContext)

	middleware := middleware.NewAppMiddleware()
	if os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	ginEngine := gin.New()

	// panic recovery
	ginEngine.Use(middleware.Recovery())

	// logger
	ginEngine.Use(middleware.Logger(io.MultiWriter(writers...)))

	ginEngine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Ticket-Token"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	superadmin_http.NewRouteSuperadmin(superadminUsecase, ginEngine, middleware)
	admin_http.NewRouteAdmin(adminUsecase, ginEngine, middleware)
	user_http.NewRouteUser(userUsecase, ginEngine, middleware)

	// Gin initialization
	ginEngine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"message": "Welcome",
		})
	})
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	logrus.Infof("Service running on port %s", port)
	ginEngine.Run(":" + port)
}
