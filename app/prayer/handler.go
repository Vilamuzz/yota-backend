package prayer

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

func NewHandler(service Service, m middleware.AppMiddleware) handler {
	return handler{service: service, middleware: m}
}

func (h *handler) RegisterRoutes(router *gin.Engine) {
	public := router.Group("/prayers")
	public.GET("/:id", h.FindPrayerByID)
	public.GET("/", h.ListPrayers)
	public.PUT("/:id/increment-count", h.IncrementPrayerCount)
	public.PUT("/:id/decrement-count", h.DecrementPrayerCount)
	public.PUT("/:id/report", h.ReportPrayer)

	protected := router.Group("/protected/prayers")
	protected.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		protected.GET("/", h.ListReportedPrayers)
		protected.DELETE("/:id", h.DeletePrayer)
	}
}

func (h *handler) IncrementPrayerCount(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.IncrementPrayerCount(ctx, id)
	c.JSON(res.Status, res)
}

func (h *handler) DecrementPrayerCount(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.DecrementPrayerCount(ctx, id)
	c.JSON(res.Status, res)
}

func (h *handler) ReportPrayer(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.ReportPrayer(ctx, id)
	c.JSON(res.Status, res)
}

// @Summary Get Prayer by ID
// @Description Get a prayer by its ID
// @Tags Prayer
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response{data=PrayerResponse}
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /prayers/{id} [get]
func (h *handler) FindPrayerByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.FindPrayerByID(ctx, id)
	c.JSON(res.Status, res)
}

// @Summary List Prayers
// @Description Get a list of prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /prayers [get]
func (h *handler) ListPrayers(c *gin.Context) {
	ctx := c.Request.Context()
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListPrayers(ctx, params)
	c.JSON(res.Status, res)
}

// @Summary List Reported Prayers
// @Description Get a list of reported prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /protected/prayers [get]
func (h *handler) ListReportedPrayers(c *gin.Context) {
	ctx := c.Request.Context()
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListReportedPrayers(ctx, params)
	c.JSON(res.Status, res)
}

// @Summary Delete Prayer
// @Description Delete a prayer by its ID
// @Tags Prayer
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /protected/prayers/{id} [delete]
func (h *handler) DeletePrayer(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.DeletePrayer(ctx, id)
	c.JSON(res.Status, res)
}
