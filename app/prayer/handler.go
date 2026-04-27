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
	router.GET("/donation-programs/:slug/prayers", h.GetPrayerList, h.middleware.AuthOptional())
	router.GET("/prayers/:id", h.GetPrayerByID, h.middleware.AuthOptional())

	publicProtected := router.Group("/prayers")
	publicProtected.Use(h.middleware.AuthRequired())
	{
		publicProtected.POST("/:id/amen", h.CreateAmenPrayer)
		publicProtected.POST("/:id/report", h.CreateReportPrayer)
	}

	admin := router.Group("/admin/prayers")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("/", h.GetReportedPrayerList)
		admin.DELETE("/:id", h.DeletePrayer)
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
// @Router /api/prayers/{id}/amen [post]
func (h *handler) CreateAmenPrayer(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	accountID := claims.AccountID
	prayerID := c.Param("id")

	res := h.service.CreatePrayerAmen(ctx, prayerID, accountID)
	c.JSON(res.Status, res)
}

// @Summary Report Prayer
// @Description Report a prayer
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Param body body ReportPrayerRequest true "Report Prayer Payload"
// @Success 200 {object} pkg.Response
// @Router /api/prayers/{id}/report [post]
func (h *handler) CreateReportPrayer(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	accountID := claims.AccountID
	var payload ReportPrayerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	prayerID := c.Param("id")
	res := h.service.CreateReportPrayer(ctx, prayerID, accountID, payload)
	c.JSON(res.Status, res)
}

// @Summary Get Prayer by ID
// @Description Get a prayer by its ID
// @Tags Prayer
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response{data=PrayerResponse}
// @Router /api/prayers/{id} [get]
func (h *handler) GetPrayerByID(c *gin.Context) {
	ctx := c.Request.Context()
	prayerID := c.Param("id")
	accountID := ""
	if accountData, exists := c.Get("user_data"); exists {
		if claims, ok := accountData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}
	res := h.service.GetPrayerByID(ctx, prayerID, accountID)
	c.JSON(res.Status, res)
}

// @Summary List Prayers
// @Description Get a list of prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Router /api/donation-programs/{slug}/prayers [get]
func (h *handler) GetPrayerList(c *gin.Context) {
	ctx := c.Request.Context()
	donationSlug := c.Param("slug")
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	accountID := ""
	if accountData, exists := c.Get("user_data"); exists {
		if claims, ok := accountData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}
	res := h.service.GetPrayerList(ctx, accountID, donationSlug, params)
	c.JSON(res.Status, res)
}

// @Summary List Reported Prayers
// @Description Get a list of reported prayers
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Router /api/admin/prayers [get]
func (h *handler) GetReportedPrayerList(c *gin.Context) {
	ctx := c.Request.Context()
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.GetReportedPrayerList(ctx, params)
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
// @Router /api/admin/prayers/{id} [delete]
func (h *handler) DeletePrayer(c *gin.Context) {
	ctx := c.Request.Context()
	prayerID := c.Param("id")
	res := h.service.DeletePrayer(ctx, prayerID)
	c.JSON(res.Status, res)
}
