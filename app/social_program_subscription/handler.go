package social_program_subscription

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, m middleware.AppMiddleware) {
	h := &handler{
		service:    s,
		middleware: m,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	// User Routes
	r.POST("/social-programs/:id/subscribe", h.middleware.RequireRoles(enum.RoleOrangTuaAsuh), h.CreateSocialProgramSubscription)
	r.PATCH("/social-programs/:id/unsubscribe", h.middleware.RequireRoles(enum.RoleOrangTuaAsuh), h.DeactivateMySocialProgramSubscription)

	me := r.Group("/social-programs/subscriptions/me")
	me.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		// me.GET("", h.GetMySocialProgramSubscriptionList)
	}

	// Admin routes
	admin := r.Group("/admin/social-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager))
	{
		admin.GET("/subscribers", h.GetSubscribers)
		admin.GET("/subscribers/:id", h.GetSubscriberByID)
		admin.GET("/:id/subscriptions", h.GetSocialProgramSubscriptionList)
		admin.GET("/accounts/:account_id/subscriptions", h.GetSocialProgramSubscriptionsByAccountID)
		admin.GET("/subscriptions/:id", h.GetSocialProgramSubscriptionByID)
		admin.POST("/:id/subscriptions", h.CreateOfflineSocialProgramSubscription)
		admin.PATCH("/subscriptions/:id/deactivate", h.DeactivateSocialProgramSubscription)
	}
}

// GetSocialProgramSubscriptionList
//
// @Summary List All Social Program Subscriptions
// @Description Retrieve a list of all social program subscriptions (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param social_program_id query string false "Filter by social program ID"
// @Param account_id query string false "Filter by account ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/subscriptions [get]
func (h *handler) GetSocialProgramSubscriptionList(c *gin.Context) {
	ctx := c.Request.Context()
	socialProgramID := c.Param("id")

	var queryParams SocialProgramSubscriptionQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSocialProgramSubscriptionList(ctx, socialProgramID, queryParams)
	c.JSON(res.Status, res)
}

// GetSocialProgramSubscriptionByID
//
// @Summary Get Social Program Subscription By ID
// @Description Retrieve a specific social program subscription by its ID
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/subscriptions/{id} [get]
func (h *handler) GetSocialProgramSubscriptionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetSocialProgramSubscriptionByID(ctx, id)
	c.JSON(res.Status, res)
}

// CreateSocialProgramSubscription
//
// @Summary Create Social Program Subscription
// @Description Subscribe to a social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Param payload body CreateSocialProgramSubscriptionRequest true "Subscription Data"
// @Success 201 {object} pkg.Response
// @Router /api/social-programs/{id}/subscribe [post]
func (h *handler) CreateSocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	socialProgramID := c.Param("id")

	res := h.service.CreateSocialProgramSubscription(ctx, claims.AccountID, socialProgramID)
	c.JSON(res.Status, res)
}

// CreateOfflineSocialProgramSubscription
//
// @Summary Create Social Program Subscription
// @Description Subscribe to a social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Param payload body CreateSocialProgramSubscriptionRequest true "Subscription Data"
// @Success 201 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/subscribe [post]
func (h *handler) CreateOfflineSocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	socialProgramID := c.Param("id")

	var req CreateSocialProgramSubscriptionOfflineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateSocialProgramSubscription(ctx, req.AccountID, socialProgramID)
	c.JSON(res.Status, res)
}

// DeactivateSocialProgramSubscription
//
// @Summary Deactivate Social Program Subscription
// @Description Deactivate a social program subscription (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/subscriptions/{id}/deactivate [patch]
func (h *handler) DeactivateSocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.DeactivateSocialProgramSubscription(ctx, id, "")
	c.JSON(res.Status, res)
}

// DeactivateMySocialProgramSubscription
//
// @Summary Deactivate My Social Program Subscription
// @Description Deactivate own social program subscription
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/subscriptions/{id}/deactivate [patch]
func (h *handler) DeactivateMySocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")

	res := h.service.DeactivateSocialProgramSubscription(ctx, id, claims.AccountID)
	c.JSON(res.Status, res)
}

// GetSubscribers
//
// @Summary List All Subscribers
// @Description Retrieve a unique list of accounts that have at least one social program subscription (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/subscribers [get]
func (h *handler) GetSubscribers(c *gin.Context) {
	ctx := c.Request.Context()

	var params pkg.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSubscribers(ctx, params)
	c.JSON(res.Status, res)
}

// GetSubscriberByID
//
// @Summary Get Subscriber By ID
// @Description Retrieve a subscriber's profile and stats by their Account ID (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Account ID (Subscriber ID)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/subscribers/{id} [get]
func (h *handler) GetSubscriberByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetSubscriberByID(ctx, id)
	c.JSON(res.Status, res)
}

// GetSocialProgramSubscriptionsByAccountID
//
// @Summary List Social Program Subscriptions by Account ID
// @Description Retrieve a list of social program subscriptions for a specific account (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param account_id path string true "Account ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/accounts/{account_id}/subscriptions [get]
func (h *handler) GetSocialProgramSubscriptionsByAccountID(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("account_id")

	var queryParams SocialProgramSubscriptionQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSocialProgramSubscriptionsByAccountID(ctx, accountID, queryParams)
	c.JSON(res.Status, res)
}
