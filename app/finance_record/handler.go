package finance_record

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
	r.GET("/finance-records/summary", h.SummaryFinanceRecord)
	r.GET("/admin/finance-records/summary", h.middleware.RequireRoles(enum.RoleFinance), h.AdminSummaryFinanceRecord)
	r.GET("/admin/finance-records/monthly-trend", h.middleware.RequireRoles(enum.RoleFinance), h.MonthlyTrend)
}

func (h *handler) SummaryFinanceRecord(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetSummary(ctx, false)
	c.JSON(res.Status, res)
}

func (h *handler) AdminSummaryFinanceRecord(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetSummary(ctx, true)
	c.JSON(res.Status, res)
}

func (h *handler) MonthlyTrend(c *gin.Context) {
	ctx := c.Request.Context()
	var params MonthlyTrendQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", nil, nil))
		return
	}
	res := h.service.GetMonthlyTrend(ctx, params)
	c.JSON(res.Status, res)
}
