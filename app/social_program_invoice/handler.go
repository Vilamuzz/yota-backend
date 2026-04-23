package social_program_invoice

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
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
	r.GET("/me/subscriptions/:id/invoices", h.GetSocialProgramInvoiceList, h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
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
// @Router /api/me/subscriptions/{id}/invoices [get]
func (h *handler) GetSocialProgramInvoiceList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramInvoiceQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

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
