package log

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
	h := &handler{service: s, middleware: m}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	protected := r.Group("/logs")
	protected.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		protected.GET("", h.ListLogs)
	}
}

// ListLogs godoc
//
// @Summary List audit logs
// @Description Retrieve a paginated list of admin audit log entries
// @Tags Logs
// @Security BearerAuth
// @Produce json
// @Param entity_type query string false "Filter by entity type (e.g. donation_program_transaction, prayer)"
// @Param entity_id   query string false "Filter by entity ID"
// @Param user_id     query string false "Filter by acting user ID"
// @Param action      query string false "Filter by action (e.g. CREATE, DELETE)"
// @Param limit       query int    false "Items per page (max 100)"
// @Param next_cursor query string false "Cursor for next page"
// @Success 200 {object} pkg.Response
// @Router /api/logs [get]
func (h *handler) ListLogs(c *gin.Context) {
	var params LogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListLogs(c.Request.Context(), params)
	c.JSON(res.Status, res)
}
