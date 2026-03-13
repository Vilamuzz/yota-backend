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
	protected := r.Group("/")
	protected.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		protected.GET("/", h.ListRecords)
	}
}

// ListRecords godoc
// @Summary List finance records
// @Description Get a list of finance records with pagination
// @Tags Finance Records
// @Accept json
// @Produce json
// @Param FundID query string false "Filter by fund ID"
// @Param SourceType query string false "Filter by source type (e.g. transaction, expense)"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
func (h *handler) ListRecords(c *gin.Context) {
	var params RecordQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.ListRecords(c.Request.Context(), params)
	c.JSON(resp.Status, resp)
}
