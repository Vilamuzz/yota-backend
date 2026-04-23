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
	me := r.Group("me/subscriptions")
	me.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		me.GET("", h.GetMySocialProgramSubscriptionList)
		// me.PATCH("/:id/status", h.UpdateSocialProgramSubscriptionStatus)
	}

	r.POST("social-programs/:id/subscribe", h.CreateSocialProgramSubscription, h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))

	// Admin routes
	admin := r.Group("/admin/social-programs/subscriptions")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager))
	{
		admin.GET("", h.GetSocialProgramSubscriptionList)
		admin.GET("/:id", h.GetSocialProgramSubscriptionByID)
	}
}

// GetMySocialProgramSubscriptionList
//
// @Summary List My Social Program Subscriptions
// @Description Retrieve a list of the authenticated user's social program subscriptions
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param social_program_id query string false "Filter by social program ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/me/subscriptions [get]
func (h *handler) GetMySocialProgramSubscriptionList(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var queryParams SocialProgramSubscriptionQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	// Enforce user filtering
	queryParams.AccountID = claims.AccountID

	res := h.service.GetSocialProgramSubscriptionList(ctx, queryParams)
	c.JSON(res.Status, res)
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
// @Router /api/admin/social-programs/subscriptions [get]
func (h *handler) GetSocialProgramSubscriptionList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramSubscriptionQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSocialProgramSubscriptionList(ctx, queryParams)
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
// @Param payload body CreateSocialProgramSubscriptionRequest true "Subscription Data"
// @Success 201 {object} pkg.Response
// @Router /api/social-programs/subscriptions [post]
func (h *handler) CreateSocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req CreateSocialProgramSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateSocialProgramSubscription(ctx, claims.AccountID, req)
	c.JSON(res.Status, res)
}

// UpdateSocialProgramSubscription
//
// @Summary Update Social Program Subscription
// @Description Update the status of a social program subscription
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param payload body UpdateSocialProgramSubscriptionRequest true "Subscription Update Data"
// @Success 200 {object} pkg.Response
// @Router /api/social-programs/subscriptions/{id} [put]
func (h *handler) UpdateSocialProgramSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req UpdateSocialProgramSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.UpdateSocialProgramSubscription(ctx, id, req)
	c.JSON(res.Status, res)
}
