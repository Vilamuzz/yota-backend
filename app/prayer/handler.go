package prayer

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

func NewHandler(r *gin.RouterGroup, service Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:    service,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	public := router.Group("/prayers")
	public.Use(h.middleware.AuthOptional())
	public.GET("", h.ListPrayers)
	public.GET("/:id", h.FindPrayerByID)

	publicProtected := router.Group("/prayers")
	publicProtected.Use(h.middleware.AuthRequired())
	{
		publicProtected.POST("/:id/amen", h.PrayerAmen)
		publicProtected.POST("/:id/report", h.ReportPrayer)

	}

	protected := router.Group("/protected/prayers")
	protected.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		protected.GET("/", h.ListReportedPrayers)
		protected.DELETE("/:id", h.DeletePrayer)
	}
}

// @Summary Toggle Amen on Prayer
// @Description Toggle amen on a prayer
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/prayers/{id}/amen [post]
func (h *handler) PrayerAmen(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	userID := claims.UserID
	prayerID := c.Param("id")

	res := h.service.PrayerAmen(ctx, prayerID, userID)
	c.JSON(res.Status, res)
}

// @Summary Report Prayer
// @Description Report a prayer
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Param payload body ReportPrayerRequest true "Report Prayer Payload"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/prayers/{id}/report [post]
func (h *handler) ReportPrayer(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	userID := claims.UserID
	var payload ReportPrayerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	prayerID := c.Param("id")
	res := h.service.CreateReportPrayer(ctx, payload, prayerID, userID)
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
// @Router /api/prayers/{id} [get]
func (h *handler) FindPrayerByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	userID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			userID = claims.UserID
		}
	}
	res := h.service.FindPrayerByID(ctx, id, userID)
	c.JSON(res.Status, res)
}

// @Summary List Prayers
// @Description Get a list of prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param donation_id query string false "Filter by donation ID"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/prayers [get]
func (h *handler) ListPrayers(c *gin.Context) {
	ctx := c.Request.Context()
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	userID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			userID = claims.UserID
		}
	}
	res := h.service.ListPrayers(ctx, params, userID)
	c.JSON(res.Status, res)
}

// @Summary List Reported Prayers
// @Description Get a list of reported prayers
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/protected/prayers [get]
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
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/protected/prayers/{id} [delete]
func (h *handler) DeletePrayer(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.DeletePrayer(ctx, id)
	c.JSON(res.Status, res)
}
