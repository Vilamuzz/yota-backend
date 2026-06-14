package finance_record

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
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
