package social_program_invoice

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
	me := r.Group("/social-programs/subscriptions/invoices/me")
	me.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		me.GET("", h.GetSocialProgramInvoiceList)
		me.GET("/:id", h.GetSocialProgramInvoiceByID)
	}

	admin := r.Group("/admin/social-programs/subscriptions/invoices")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleFinance))
	{
		admin.GET("", h.GetSocialProgramInvoiceList)
		admin.GET("/subscription/:id", h.GetSocialProgramInvoiceListBySubscriptionID)
	}
}

// GetSocialProgramInvoiceList
//
// @Summary List Social Program Invoices
// @Description Retrieve a paginated list of social program invoices
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param subscription_id query string false "Filter by subscription ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/subscriptions/invoices/me [get]
func (h *handler) GetSocialProgramInvoiceList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramInvoiceQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
		return
	}

	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			queryParams.AccountID = claims.AccountID
		}
	}

	res := h.service.GetSocialProgramInvoiceList(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetSocialProgramInvoiceListBySubscriptionID
//
// @Summary List Social Program Invoices by Subscription ID
// @Description Retrieve a paginated list of social program invoices for a specific subscription
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/subscriptions/invoices/subscription/{id} [get]
func (h *handler) GetSocialProgramInvoiceListBySubscriptionID(c *gin.Context) {
	ctx := c.Request.Context()
	subscriptionID := c.Param("id")

	var queryParams SocialProgramInvoiceQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
		return
	}

	queryParams.SubscriptionID = subscriptionID

	res := h.service.GetSocialProgramInvoiceList(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetSocialProgramInvoiceByID
//
// @Summary Get Social Program Invoice by ID
// @Description Get detailed information of a specific social program invoice
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/subscriptions/invoices/{id} [get]
func (h *handler) GetSocialProgramInvoiceByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetSocialProgramInvoiceByID(ctx, id)
	c.JSON(res.Status, res)
}
