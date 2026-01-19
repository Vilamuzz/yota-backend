package router

import (
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/gin-gonic/gin"
)

type Router struct {
	AuthHandler *auth.Handler
	Middleware  *middleware.AppMiddleware
}

func NewRouter(authHandler *auth.Handler, appMiddleware *middleware.AppMiddleware) *Router {
	return &Router{
		AuthHandler: authHandler,
		Middleware:  appMiddleware,
	}
}

const (
	RoleUser               = "user"
	RoleChairman           = "chairman"
	RoleEventManager       = "event_manager"
	RoleFinance            = "finance"
	RoleAmbulanceManager   = "ambulance_manager"
	RolePublicationManager = "publication_manager"
	RoleCaseManager        = "case_manager"
	RoleSuperadmin         = "superadmin"
)

func (r *Router) SetupRoutes(engine *gin.Engine) {
	api := engine.Group("/api")

	auth.RegisterRoutes(api, r.AuthHandler)

	// Protected routes
	r.setupProtectedRoutes(api)
}

func (r *Router) setupProtectedRoutes(api *gin.RouterGroup) {
	// User-related routes (any authenticated user can access their own profile)
	userRoutes := api.Group("/users")
	userRoutes.Use(r.Middleware.AuthRequired())
	{
		// userRoutes.GET("/profile", userHandler.GetProfile)
		// userRoutes.PUT("/profile", userHandler.UpdateProfile)
		// userRoutes.GET("/me", userHandler.GetCurrentUser)
	}

	// Event management routes
	eventRoutes := api.Group("/events")
	{
		// Public events (no auth required)
		// eventRoutes.GET("/", eventHandler.GetPublicEvents)
		// eventRoutes.GET("/:id", eventHandler.GetEventDetails)

		// Protected routes for event management
		eventManagement := eventRoutes.Group("/")
		eventManagement.Use(r.Middleware.RequireRoles(RoleEventManager, RoleChairman, RoleSuperadmin))
		{
			// eventManagement.POST("/", eventHandler.CreateEvent)
			// eventManagement.PUT("/:id", eventHandler.UpdateEvent)
			// eventManagement.DELETE("/:id", eventHandler.DeleteEvent)
			// eventManagement.POST("/:id/publish", eventHandler.PublishEvent)
		}
	}

	// Finance management routes
	financeRoutes := api.Group("/finance")
	financeRoutes.Use(r.Middleware.RequireRoles(RoleFinance, RoleChairman, RoleSuperadmin))
	{
		// financeRoutes.GET("/reports", financeHandler.GetReports)
		// financeRoutes.POST("/transactions", financeHandler.CreateTransaction)
		// financeRoutes.GET("/budget", financeHandler.GetBudget)
		// financeRoutes.PUT("/budget", financeHandler.UpdateBudget)
	}

	// Ambulance management routes
	ambulanceRoutes := api.Group("/ambulance")
	ambulanceRoutes.Use(r.Middleware.RequireRoles(RoleAmbulanceManager, RoleChairman, RoleSuperadmin))
	{
		// ambulanceRoutes.GET("/vehicles", ambulanceHandler.GetVehicles)
		// ambulanceRoutes.POST("/vehicles", ambulanceHandler.CreateVehicle)
		// ambulanceRoutes.GET("/requests", ambulanceHandler.GetRequests)
		// ambulanceRoutes.PUT("/requests/:id/assign", ambulanceHandler.AssignRequest)
	}

	// Publications routes
	publicationRoutes := api.Group("/publications")
	{
		// Public access to published content
		// publicationRoutes.GET("/", publicationHandler.GetPublishedArticles)
		// publicationRoutes.GET("/:id", publicationHandler.GetArticle)

		// Management routes
		publicationManagement := publicationRoutes.Group("/manage")
		publicationManagement.Use(r.Middleware.RequireRoles(RolePublicationManager, RoleChairman, RoleSuperadmin))
		{
			// publicationManagement.GET("/", publicationHandler.GetAllArticles)
			// publicationManagement.POST("/", publicationHandler.CreateArticle)
			// publicationManagement.PUT("/:id", publicationHandler.UpdateArticle)
			// publicationManagement.DELETE("/:id", publicationHandler.DeleteArticle)
			// publicationManagement.POST("/:id/publish", publicationHandler.PublishArticle)
		}
	}

	// Case management routes
	caseRoutes := api.Group("/cases")
	caseRoutes.Use(r.Middleware.RequireRoles(RoleCaseManager, RoleChairman, RoleSuperadmin))
	{
		// caseRoutes.GET("/", caseHandler.GetCases)
		// caseRoutes.POST("/", caseHandler.CreateCase)
		// caseRoutes.GET("/:id", caseHandler.GetCase)
		// caseRoutes.PUT("/:id", caseHandler.UpdateCase)
		// caseRoutes.PUT("/:id/status", caseHandler.UpdateCaseStatus)
	}

	// Administrative routes
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(r.Middleware.RequireRoles(RoleChairman, RoleSuperadmin))
	{
		// adminRoutes.GET("/users", userHandler.GetAllUsers)
		// adminRoutes.PUT("/users/:id/role", userHandler.UpdateUserRole)
		// adminRoutes.PUT("/users/:id/status", userHandler.UpdateUserStatus)
		// adminRoutes.GET("/dashboard", adminHandler.GetDashboard)
		// adminRoutes.GET("/analytics", adminHandler.GetAnalytics)
	}

	// System management (superadmin only)
	systemRoutes := api.Group("/system")
	systemRoutes.Use(r.Middleware.RequireRoles(RoleSuperadmin))
	{
		// systemRoutes.GET("/health", systemHandler.HealthCheck)
		// systemRoutes.POST("/backup", systemHandler.CreateBackup)
		// systemRoutes.GET("/logs", systemHandler.GetSystemLogs)
		// systemRoutes.PUT("/settings", systemHandler.UpdateSystemSettings)
	}
}
