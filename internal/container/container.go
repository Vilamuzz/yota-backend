package container

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/app/ambulance_request"
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/donation_expense"
	"github.com/Vilamuzz/yota-backend/app/donation_transaction"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/gallery"
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/internal/scheduler"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	redis_pkg "github.com/Vilamuzz/yota-backend/pkg/redis"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	// Infrastructure
	DB             *gorm.DB
	RedisClient    *redis_pkg.Client
	S3Client       s3_pkg.Client
	MidtransClient payment_pkg.Client
	Timeout        time.Duration

	// Repositories
	UserRepo                user.Repository
	AuthRepo                auth.Repository
	DonationRepo            donation.Repository
	NewsRepo                news.Repository
	GalleryRepo             gallery.Repository
	MediaRepo               media.Repository
	PrayerRepo              prayer.Repository
	TransactionDonationRepo donation_transaction.Repository
	DonationExpenseRepo     donation_expense.Repository
	FinanceRecordRepo       finance_record.Repository
	AmbulanceRepo           ambulance.Repository
	AmbulanceHistoryRepo    ambulance_history.Repository
	AmbulanceRequestRepo    ambulance_request.Repository

	// Services
	AuthService                auth.Service
	UserService                user.Service
	DonationService            donation.Service
	NewsService                news.Service
	GalleryService             gallery.Service
	MediaService               media.Service
	TransactionDonationService donation_transaction.Service
	PrayerService              prayer.Service
	DonationExpenseService     donation_expense.Service
	FinanceRecordService       finance_record.Service
	AmbulanceService           ambulance.Service
	AmbulanceHistoryService    ambulance_history.Service
	AmbulanceRequestService    ambulance_request.Service

	// Middleware
	Middleware *middleware.AppMiddleware

	// Scheduler
	Scheduler *scheduler.Scheduler
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

	// Initialize scheduler
	c.initScheduler()

	// Cleanup function
	cleanup := func() {
		if c.Scheduler != nil {
			c.Scheduler.Stop()
		}
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

	// S3-compatible client (RustFS)
	s3Client := config.ConnectS3()
	c.S3Client = s3_pkg.NewClient(s3Client)

	// Timeout
	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	c.Timeout = time.Duration(timeout) * time.Second

	// Midtrans
	c.MidtransClient = payment_pkg.NewClient()

	return nil
}

func (c *Container) initRepositories() {
	c.UserRepo = user.NewRepository(c.DB)
	c.AuthRepo = auth.NewRepository(c.DB)
	c.DonationRepo = donation.NewRepository(c.DB)
	c.NewsRepo = news.NewRepository(c.DB)
	c.GalleryRepo = gallery.NewRepository(c.DB)
	c.MediaRepo = media.NewRepository(c.DB)
	c.PrayerRepo = prayer.NewRepository(c.DB)
	c.TransactionDonationRepo = donation_transaction.NewRepository(c.DB)
	c.DonationExpenseRepo = donation_expense.NewRepository(c.DB)
	c.FinanceRecordRepo = finance_record.NewRepository(c.DB)
	c.AmbulanceRepo = ambulance.NewRepository(c.DB)
	c.AmbulanceHistoryRepo = ambulance_history.NewRepository(c.DB)
	c.AmbulanceRequestRepo = ambulance_request.NewRepository(c.DB)
}

func (c *Container) initServices() {
	c.AuthService = auth.NewService(c.AuthRepo, c.UserRepo, c.Timeout)
	c.UserService = user.NewService(c.UserRepo, c.Timeout)
	c.DonationService = donation.NewService(c.DonationRepo, c.S3Client, c.Timeout)
	c.NewsService = news.NewService(c.NewsRepo, c.Timeout)
	c.MediaService = media.NewService(c.MediaRepo, c.S3Client)
	c.GalleryService = gallery.NewService(c.GalleryRepo, c.MediaService, c.Timeout)
	c.TransactionDonationService = donation_transaction.NewService(c.TransactionDonationRepo, c.UserRepo, c.DonationRepo, c.PrayerRepo, c.FinanceRecordRepo, c.MidtransClient, c.Timeout)
	c.PrayerService = prayer.NewService(c.PrayerRepo, c.Timeout)
	c.DonationExpenseService = donation_expense.NewService(c.DonationExpenseRepo, c.FinanceRecordRepo, c.DonationRepo, c.S3Client, c.Timeout)
	c.FinanceRecordService = finance_record.NewService(c.FinanceRecordRepo, c.Timeout)
	c.AmbulanceService = ambulance.NewService(c.AmbulanceRepo, c.S3Client, c.Timeout)
	c.AmbulanceHistoryService = ambulance_history.NewService(c.AmbulanceHistoryRepo, c.AmbulanceRepo, c.Timeout)
	c.AmbulanceRequestService = ambulance_request.NewService(c.AmbulanceRequestRepo, c.Timeout)
}

func (c *Container) initMiddleware() {
	var redisClient *redis.Client
	if c.RedisClient != nil {
		redisClient = c.RedisClient.GetClient()
	}
	c.Middleware = middleware.NewAppMiddleware(redisClient)
}

func (c *Container) initScheduler() {
	c.Scheduler = scheduler.New()

	// Update expired donations to 'complete' every midnight
	c.Scheduler.Add("0 0 * * *", "update-expired-donations", func() {
		if err := c.DonationService.UpdateExpiredDonations(context.Background()); err != nil {
			// error is logged inside the scheduler wrapper; log detail here
			_ = err
		}
	})
}

// RegisterHandlers registers all handlers with their routes
func (c *Container) RegisterHandlers(router *gin.RouterGroup) {
	auth.NewHandler(router, c.AuthService, c.UserService, *c.Middleware)
	user.NewHandler(router, c.UserService, *c.Middleware)
	donation.NewHandler(router, c.DonationService, *c.Middleware)
	news.NewHandler(router, c.NewsService, c.S3Client, *c.Middleware)
	gallery.NewHandler(router, c.GalleryService, c.MediaService, *c.Middleware)
	donation_transaction.NewHandler(router, c.TransactionDonationService, *c.Middleware)
	prayer.NewHandler(router, c.PrayerService, *c.Middleware)
	donation_expense.NewHandler(router, c.DonationExpenseService, *c.Middleware)
	finance_record.NewHandler(router, c.FinanceRecordService, *c.Middleware)
	ambulance.NewHandler(router, c.AmbulanceService, *c.Middleware)
	ambulance_history.NewHandler(router, c.AmbulanceHistoryService, *c.Middleware)
	ambulance_request.NewHandler(router, c.AmbulanceRequestService, *c.Middleware)
}
