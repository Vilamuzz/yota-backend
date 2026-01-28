package router

import (
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/gin-gonic/gin"
)

type Router struct {
	AuthHandler     *auth.Handler
	DonationHandler *donation.Handler
	Middleware      *middleware.AppMiddleware
}

func NewRouter(authHandler *auth.Handler, donationHandler *donation.Handler, appMiddleware *middleware.AppMiddleware) *Router {
	return &Router{
		AuthHandler:     authHandler,
		DonationHandler: donationHandler,
		Middleware:      appMiddleware,
	}
}

const (
	RoleUser               = "user"
	RoleChairman           = "chairman"
	RoleSocialManager      = "social_manager"
	RoleFinance            = "finance"
	RoleAmbulanceManager   = "ambulance_manager"
	RolePublicationManager = "publication_manager"
	RoleSuperadmin         = "superadmin"
)

func (r *Router) Routes(engine *gin.Engine) {
	api := engine.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", r.AuthHandler.Register)
		auth.POST("/login", r.AuthHandler.Login)
		auth.POST("/forget-password", r.AuthHandler.ForgetPassword)
		auth.POST("/reset-password", r.AuthHandler.ResetPassword)
	}

	// Protected routes
	r.protectedRoutes(api)
}

func (r *Router) protectedRoutes(api *gin.RouterGroup) {
	// User-related routes (any authenticated user can access their own profile)
	userRoutes := api.Group("/users")
	userRoutes.Use(r.Middleware.AuthRequired())
	{
		// userRoutes.GET("/profile", userHandler.GetProfile)
		// userRoutes.PUT("/profile", userHandler.UpdateProfile)
		// userRoutes.GET("/me", userHandler.GetCurrentUser)
	}

	// Finance management routes
	financeRoutes := api.Group("/finance")
	financeRoutes.Use(r.Middleware.RequireRoles(RoleFinance))
	{

	}
	donationRoutes := financeRoutes.Group("/donations")
	{
		donationRoutes.GET("/", r.DonationHandler.GetAllDonations)
	}
}
