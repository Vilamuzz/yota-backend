package container

import (
	"os"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/config"
	postgre_pkg "github.com/Vilamuzz/yota-backend/pkg/postgre"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Container struct {
	// Infrastructure
	DB      *gorm.DB
	Timeout time.Duration

	// Repositories
	UserRepo     user.Repository
	AuthRepo     auth.Repository
	DonationRepo donation.Repository
	NewsRepo     news.Repository

	// Services
	AuthService     auth.Service
	DonationService donation.Service
	NewsService     news.Service

	// Middleware
	Middleware *middleware.AppMiddleware
}

func NewContainer() (*Container, func(), error) {
	c := &Container{}

	// Initialize infrastructure
	if err := c.initInfrastructure(); err != nil {
		return nil, nil, err
	}

	// Initialize repositories
	c.initRepositories()

	// Initialize services
	c.initServices()

	// Initialize middleware
	c.initMiddleware()

	// Cleanup function
	cleanup := func() {
		if c.DB != nil {
			sqlDB, _ := c.DB.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	}

	return c, cleanup, nil
}

func (c *Container) initInfrastructure() error {
	// Database
	db := config.ConnectDB()
	postgre_pkg.AutoMigrateDB(db, postgre_pkg.GetAllModels()...)
	c.DB = db

	// Timeout
	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	c.Timeout = time.Duration(timeout) * time.Second

	return nil
}

func (c *Container) initRepositories() {
	c.UserRepo = user.NewRepository(c.DB)
	c.AuthRepo = auth.NewRepository(c.DB)
	c.DonationRepo = donation.NewRepository(c.DB)
	c.NewsRepo = news.NewRepository(c.DB)
}

func (c *Container) initServices() {
	c.AuthService = auth.NewService(c.UserRepo, c.AuthRepo, c.Timeout)
	c.DonationService = donation.NewService(c.DonationRepo, c.Timeout)
	c.NewsService = news.NewService(c.NewsRepo, c.Timeout)
}

func (c *Container) initMiddleware() {
	c.Middleware = middleware.NewAppMiddleware()
}

// RegisterHandlers registers all handlers with their routes
func (c *Container) RegisterHandlers(router *gin.RouterGroup) {
	auth.NewHandler(c.AuthService, router, *c.Middleware)
	donation.NewHandler(c.DonationService, router, *c.Middleware)
	news.NewHandler(c.NewsService, router, *c.Middleware)
}
