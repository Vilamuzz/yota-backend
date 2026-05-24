package finance_record

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
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
}

func (h *handler) SummaryFinanceRecord(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetSummary(ctx)
	c.JSON(res.Status, res)
}
