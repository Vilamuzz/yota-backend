package container

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/config"
	redis_pkg "github.com/Vilamuzz/yota-backend/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	// Infrastructure
	DB          *gorm.DB
	RedisClient *redis_pkg.Client
	Timeout     time.Duration

	// Repositories
	UserRepo     user.Repository
	AuthRepo     auth.Repository
	DonationRepo donation.Repository
	NewsRepo     news.Repository

	// Services
	AuthService     auth.Service
	UserService     user.Service
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
		if c.RedisClient != nil {
			c.RedisClient.Close()
		}
	}

	return c, cleanup, nil
}

func (c *Container) initInfrastructure() error {
	// Database
	db := config.ConnectDB()
	c.DB = db

	// Redis
	redisClient, err := redis_pkg.NewClient()
	if err != nil {
		fmt.Printf("Warning: Redis connection failed: %v. Falling back to memory store.\n", err)
		c.RedisClient = nil
	} else {
		c.RedisClient = redisClient
	}

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
	c.UserService = user.NewService(c.UserRepo, c.Timeout)
	c.DonationService = donation.NewService(c.DonationRepo, c.Timeout)
	c.NewsService = news.NewService(c.NewsRepo, c.Timeout)
}

func (c *Container) initMiddleware() {
	var redisClient *redis.Client
	if c.RedisClient != nil {
		redisClient = c.RedisClient.GetClient()
	}
	c.Middleware = middleware.NewAppMiddleware(redisClient)
}

// RegisterHandlers registers all handlers with their routes
func (c *Container) RegisterHandlers(router *gin.RouterGroup) {
	auth.NewHandler(router, c.AuthService, c.UserService, *c.Middleware)
	user.NewHandler(router, c.UserService, *c.Middleware)
	donation.NewHandler(router, c.DonationService, *c.Middleware)
	news.NewHandler(router, c.NewsService, *c.Middleware)
}
