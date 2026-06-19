package container

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/app/ambulance_service_request"
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/backup"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/donation_program_expense"
	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/app/foster_children_candidate"
	"github.com/Vilamuzz/yota-backend/app/foster_children_expense"
	"github.com/Vilamuzz/yota-backend/app/foster_children_transaction"
	"github.com/Vilamuzz/yota-backend/app/foundation_profile"
	"github.com/Vilamuzz/yota-backend/app/gallery"
	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/news_comment"
	"github.com/Vilamuzz/yota-backend/app/payment"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/social_program_expense"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/Vilamuzz/yota-backend/app/social_program_transaction"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/internal/scheduler"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	redis_pkg "github.com/Vilamuzz/yota-backend/pkg/redis"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	// Infrastructure
	DB             *gorm.DB
	RedisClient    *redis_pkg.Client
	S3Client       s3_pkg.Client
	MinioClient    *minio.Client
	MidtransClient payment_pkg.Client
	Timeout        time.Duration

	// Repositories
	AccountRepo                   account.Repository
	AuthRepo                      auth.Repository
	DonationRepo                  donation_program.Repository
	NewsRepo                      news.Repository
	NewsCommentRepo               news_comment.Repository
	FoundationProfileRepo         foundation_profile.Repository
	GalleryRepo                   gallery.Repository
	MediaRepo                     media.Repository
	PrayerRepo                    prayer.Repository
	TransactionDonationRepo       donation_program_transaction.Repository
	DonationExpenseRepo           donation_program_expense.Repository
	FinanceRecordRepo             finance_record.Repository
	AmbulanceRepo                 ambulance.Repository
	AmbulanceHistoryRepo          ambulance_history.Repository
	AmbulanceServiceRequestRepo   ambulance_service_request.Repository
	FosterChildrenRepo            foster_children.Repository
	FosterChildrenCandidateRepo   foster_children_candidate.Repository
	FosterChildrenExpenseRepo     foster_children_expense.Repository
	FosterChildrenTransactionRepo foster_children_transaction.Repository
	SocialProgramRepo             social_program.Repository
	SocialProgramExpenseRepo      social_program_expense.Repository
	SocialProgramInvoiceRepo      social_program_invoice.Repository
	SocialProgramSubscriptionRepo social_program_subscription.Repository
	SocialProgramTransactionRepo  social_program_transaction.Repository
	LogRepo                       app_log.Repository

	// Services
	AuthService                      auth.Service
	AccountService                   account.Service
	DonationService                  donation_program.Service
	NewsService                      news.Service
	NewsCommentService               news_comment.Service
	FoundationProfileService         foundation_profile.Service
	GalleryService                   gallery.Service
	MediaService                     media.Service
	TransactionDonationService       donation_program_transaction.Service
	PrayerService                    prayer.Service
	DonationExpenseService           donation_program_expense.Service
	FinanceRecordService             finance_record.Service
	AmbulanceService                 ambulance.Service
	AmbulanceHistoryService          ambulance_history.Service
	AmbulanceServiceRequestService   ambulance_service_request.Service
	FosterChildrenService            foster_children.Service
	FosterChildrenCandidateService   foster_children_candidate.Service
	FosterChildrenExpenseService     foster_children_expense.Service
	FosterChildrenTransactionService foster_children_transaction.Service
	SocialProgramService             social_program.Service
	SocialProgramExpenseService      social_program_expense.Service
	SocialProgramInvoiceService      social_program_invoice.Service
	SocialProgramSubscriptionService social_program_subscription.Service
	SocialProgramTransactionService  social_program_transaction.Service
	LogService                       app_log.Service
	BackupService                    backup.Service
	BackupRepo                       backup.Repository

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

	// S3-compatible client
	minioClient := config.ConnectS3()
	c.MinioClient = minioClient
	c.S3Client = s3_pkg.NewClient(minioClient)

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
	c.AccountRepo = account.NewRepository(c.DB)
	c.AuthRepo = auth.NewRepository(c.DB)
	c.FinanceRecordRepo = finance_record.NewRepository(c.DB)
	c.DonationRepo = donation_program.NewRepository(c.DB)
	c.NewsRepo = news.NewRepository(c.DB)
	c.NewsCommentRepo = news_comment.NewRepository(c.DB)
	c.FoundationProfileRepo = foundation_profile.NewRepository(c.DB)
	c.GalleryRepo = gallery.NewRepository(c.DB)
	c.MediaRepo = media.NewRepository(c.DB)
	c.PrayerRepo = prayer.NewRepository(c.DB)
	c.TransactionDonationRepo = donation_program_transaction.NewRepository(c.DB)
	c.DonationExpenseRepo = donation_program_expense.NewRepository(c.DB)
	c.AmbulanceRepo = ambulance.NewRepository(c.DB)
	c.AmbulanceHistoryRepo = ambulance_history.NewRepository(c.DB)
	c.AmbulanceServiceRequestRepo = ambulance_service_request.NewRepository(c.DB)
	c.FosterChildrenRepo = foster_children.NewRepository(c.DB)
	c.FosterChildrenCandidateRepo = foster_children_candidate.NewRepository(c.DB)
	c.FosterChildrenExpenseRepo = foster_children_expense.NewRepository(c.DB)
	c.FosterChildrenTransactionRepo = foster_children_transaction.NewRepository(c.DB)
	c.SocialProgramRepo = social_program.NewRepository(c.DB)
	c.SocialProgramExpenseRepo = social_program_expense.NewRepository(c.DB)
	c.SocialProgramInvoiceRepo = social_program_invoice.NewRepository(c.DB)
	c.SocialProgramSubscriptionRepo = social_program_subscription.NewRepository(c.DB)
	c.SocialProgramTransactionRepo = social_program_transaction.NewRepository(c.DB)
	c.LogRepo = app_log.NewRepository(c.DB)
	c.BackupRepo = backup.NewRepository(c.DB)
}

func (c *Container) initServices() {
	c.LogService = app_log.NewService(c.LogRepo, c.Timeout)
	c.AuthService = auth.NewService(c.AuthRepo, c.AccountRepo, c.Timeout)
	c.AccountService = account.NewService(c.AccountRepo, c.Timeout, c.S3Client)
	c.FinanceRecordService = finance_record.NewService(c.FinanceRecordRepo, c.Timeout)
	c.DonationService = donation_program.NewService(c.DonationRepo, c.LogService, c.S3Client, c.Timeout)
	c.MediaService = media.NewService(c.MediaRepo, c.S3Client)
	c.NewsService = news.NewService(c.NewsRepo, c.LogService, c.S3Client, c.MediaService, c.Timeout)
	c.NewsCommentService = news_comment.NewService(c.NewsCommentRepo, c.NewsRepo, c.Timeout)
	c.FoundationProfileService = foundation_profile.NewService(c.FoundationProfileRepo, c.LogService, c.S3Client, c.Timeout)
	c.GalleryService = gallery.NewService(c.GalleryRepo, c.LogService, c.S3Client, c.MediaService, c.Timeout)
	c.TransactionDonationService = donation_program_transaction.NewService(c.TransactionDonationRepo, c.AccountRepo, c.DonationRepo, c.PrayerRepo, c.FinanceRecordRepo, c.MidtransClient, c.LogService, c.Timeout)
	c.PrayerService = prayer.NewService(c.PrayerRepo, c.DonationRepo, c.Timeout)
	c.DonationExpenseService = donation_program_expense.NewService(c.DonationExpenseRepo, c.FinanceRecordRepo, c.DonationRepo, c.S3Client, c.LogService, c.Timeout)
	c.AmbulanceService = ambulance.NewService(c.AmbulanceRepo, c.S3Client, c.Timeout)
	c.AmbulanceHistoryService = ambulance_history.NewService(c.AmbulanceHistoryRepo, c.AmbulanceRepo, c.Timeout)
	c.AmbulanceServiceRequestService = ambulance_service_request.NewService(c.AmbulanceServiceRequestRepo, c.AmbulanceRepo, c.AmbulanceHistoryRepo, c.Timeout, c.S3Client)
	c.FosterChildrenService = foster_children.NewService(c.FosterChildrenRepo, c.LogService, c.S3Client, c.Timeout)
	c.FosterChildrenCandidateService = foster_children_candidate.NewService(c.FosterChildrenCandidateRepo, c.FosterChildrenRepo, c.LogService, c.S3Client, c.Timeout)
	c.FosterChildrenExpenseService = foster_children_expense.NewService(c.FosterChildrenExpenseRepo, c.FinanceRecordRepo, c.FosterChildrenRepo, c.S3Client, c.LogService, c.Timeout)
	c.FosterChildrenTransactionService = foster_children_transaction.NewService(c.FosterChildrenTransactionRepo, c.AccountRepo, c.FosterChildrenRepo, c.FinanceRecordRepo, c.MidtransClient, c.LogService, c.Timeout)
	c.SocialProgramService = social_program.NewService(c.SocialProgramRepo, c.LogService, c.S3Client, c.Timeout)
	c.SocialProgramExpenseService = social_program_expense.NewService(c.SocialProgramExpenseRepo, c.FinanceRecordRepo, c.SocialProgramRepo, c.S3Client, c.LogService, c.Timeout)
	c.SocialProgramInvoiceService = social_program_invoice.NewService(c.SocialProgramInvoiceRepo, c.SocialProgramSubscriptionRepo, c.Timeout)
	c.SocialProgramSubscriptionService = social_program_subscription.NewService(c.SocialProgramSubscriptionRepo, c.SocialProgramRepo, c.Timeout)
	c.SocialProgramTransactionService = social_program_transaction.NewService(c.SocialProgramTransactionRepo, c.AccountRepo, c.SocialProgramSubscriptionRepo, c.SocialProgramInvoiceRepo, c.FinanceRecordRepo, c.MidtransClient, c.LogService, c.Timeout)
	c.BackupService = backup.NewService(c.BackupRepo, c.MinioClient, c.Timeout)
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
		if err := c.DonationService.UpdateExpiredDonationProgram(context.Background()); err != nil {
			_ = err
		}
	})

	// Generate monthly invoices for social programs every midnight
	c.Scheduler.Add("0 0 * * *", "generate-monthly-invoices", func() {
		if err := c.SocialProgramInvoiceService.GenerateMonthlyInvoices(context.Background()); err != nil {
			_ = err
		}
	})

	// Mark overdue invoices every midnight at 00:05
	c.Scheduler.Add("5 0 * * *", "mark-overdue-invoices", func() {
		if err := c.SocialProgramInvoiceService.MarkOverdueInvoices(context.Background()); err != nil {
			_ = err
		}
	})

	// Create database backup daily at 2 AM
	c.Scheduler.Add("0 2 * * *", "database-backup", func() {
		_ = c.BackupService.CreateBackup(context.Background())
	})

	// Cleanup old backups daily at 3 AM (keep backups for last 7 days)
	c.Scheduler.Add("0 3 * * *", "backup-cleanup", func() {
		retentionDays := 7
		_ = c.BackupService.CleanupOldBackups(context.Background(), retentionDays)
	})
}

// RegisterHandlers registers all handlers with their routes
func (c *Container) RegisterHandlers(router *gin.RouterGroup) {
	auth.NewHandler(router, c.AuthService, c.AccountService, *c.Middleware)
	account.NewHandler(router, c.AccountService, *c.Middleware)
	finance_record.NewHandler(router, c.FinanceRecordService, *c.Middleware)
	donation_program.NewHandler(router, c.DonationService, *c.Middleware)
	donation_program_transaction.NewHandler(router, c.TransactionDonationService, *c.Middleware)
	donation_program_expense.NewHandler(router, c.DonationExpenseService, *c.Middleware)
	prayer.NewHandler(router, c.PrayerService, *c.Middleware)
	news.NewHandler(router, c.NewsService, *c.Middleware)
	news_comment.NewHandler(router, c.NewsCommentService, *c.Middleware)
	foundation_profile.NewHandler(router, c.FoundationProfileService, *c.Middleware)
	gallery.NewHandler(router, c.GalleryService, c.MediaService, *c.Middleware)
	ambulance.NewHandler(router, c.AmbulanceService, *c.Middleware)
	ambulance_history.NewHandler(router, c.AmbulanceHistoryService, *c.Middleware)
	ambulance_service_request.NewHandler(router, c.AmbulanceServiceRequestService, *c.Middleware)
	foster_children.NewHandler(router, c.FosterChildrenService, *c.Middleware)
	foster_children_candidate.NewHandler(router, c.FosterChildrenCandidateService, *c.Middleware)
	foster_children_expense.NewHandler(router, c.FosterChildrenExpenseService, *c.Middleware)
	foster_children_transaction.NewHandler(router, c.FosterChildrenTransactionService, *c.Middleware)
	social_program.NewHandler(router, c.SocialProgramService, *c.Middleware)
	social_program_expense.NewHandler(router, c.SocialProgramExpenseService, *c.Middleware)
	social_program_invoice.NewHandler(router, c.SocialProgramInvoiceService, *c.Middleware)
	social_program_subscription.NewHandler(router, c.SocialProgramSubscriptionService, *c.Middleware)
	social_program_transaction.NewHandler(router, c.SocialProgramTransactionService, *c.Middleware)
	app_log.NewHandler(router, c.LogService, *c.Middleware)
	backup.NewHandler(router, c.BackupService, *c.Middleware)

	// Payment Webhooks
	paymentGroup := router.Group("/webhooks")
	payment.NewHandler(paymentGroup, c.TransactionDonationService, c.SocialProgramTransactionService, c.FosterChildrenTransactionService)
}
