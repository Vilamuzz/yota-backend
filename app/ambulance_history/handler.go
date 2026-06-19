package ambulance_history

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
	r.GET("/ambulances/:id/history", h.ListAmbulanceHistory)
	r.GET("/ambulances/:id/history/summary", h.AmbulanceHistorySummary)

	ambulanceManager := r.Group("/admin/ambulances/history")
	ambulanceManager.Use(h.middleware.RequireRoles(enum.RoleAmbulanceManager))
	{
		ambulanceManager.GET("/:id", h.AdminListAmbulanceHistory)
		ambulanceManager.POST("", h.CreateAmbulanceHistory)
		ambulanceManager.PUT("/:id", h.UpdateAmbulanceHistory)
		ambulanceManager.DELETE("/:id", h.DeleteAmbulanceHistory)
		ambulanceManager.GET("/summary", h.AllHistorySummary)
		ambulanceManager.GET("/monthly-trend", h.HistoryMonthlyTrend)
	}

	ambulanceDriver := r.Group("/admin/ambulances/history/driver")
	ambulanceDriver.Use(h.middleware.RequireRoles(enum.RoleAmbulanceDriver))
	{
		ambulanceDriver.GET("", h.DriverListAmbulanceHistory)
		ambulanceDriver.POST("", h.CreateAmbulanceHistory)
		ambulanceDriver.PUT("/:id", h.UpdateAmbulanceHistory)
		ambulanceDriver.DELETE("/:id", h.DeleteAmbulanceHistory)
		ambulanceDriver.GET("/summary", h.DriverHistorySummary)
		ambulanceDriver.GET("/monthly-trend", h.DriverHistoryMonthlyTrend)
	}
}

// ListAmbulanceHistory godoc
// @Summary List ambulance history
// @Description Get a list of ambulance history records with pagination
// @Tags Ambulance History
// @Accept json
// @Produce json
// @Param ambulance_id query int false "Filter by ambulance ID"
// @Param service_category query string false "Filter by service category"
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-history [get]
func (h *handler) ListAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams AmbulanceHistoryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	queryParams.AmbulanceID = c.Param("id")

	res := h.service.ListAmbulanceHistory(ctx, queryParams)
	c.JSON(res.Status, res)
}

// AdminListAmbulanceHistory godoc
// @Summary List ambulance history for admin
// @Description Get a list of ambulance history records with pagination for admin
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance ID"
// @Param service_category query string false "Filter by service category"
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/{id} [get]
func (h *handler) AdminListAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams AmbulanceHistoryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	queryParams.AmbulanceID = c.Param("id")

	res := h.service.AdminListAmbulanceHistory(ctx, queryParams)
	c.JSON(res.Status, res)
}

// DriverListAmbulanceHistory godoc
// @Summary List ambulance history for driver
// @Description Get a list of ambulance history records with pagination for driver
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param service_category query string false "Filter by service category"
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/driver [get]
func (h *handler) DriverListAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var queryParams AmbulanceHistoryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	queryParams.DriverID = claims.AccountID

	res := h.service.DriverListAmbulanceHistory(ctx, queryParams)
	c.JSON(res.Status, res)
}

// HistoryMonthlyTrend godoc
// @Summary Get ambulance history monthly trend
// @Description Retrieve aggregated monthly trend of ambulance history (social_service, mortuary_service, patient_service, emergency_service, other_service) for a given year (admin only)
// @Tags Ambulance History
// @Security BearerAuth
// @Produce json
// @Param year query string false "Filter by year"
// @Success 200 {object} pkg.Response{data=HistoryMonthlyTrendRecord}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/monthly-trend [get]
func (h *handler) HistoryMonthlyTrend(c *gin.Context) {
	ctx := c.Request.Context()

	var params MonthlyTrendQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.HistoryMonthlyTrend(ctx, params)
	c.JSON(res.Status, res)
}

// AllHistorySummary godoc
// @Summary Get all ambulance history summary for admin
// @Description Returns total service counts grouped by category for all ambulances.
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} pkg.Response{data=SummaryResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/summary [get]
func (h *handler) AllHistorySummary(c *gin.Context) {
	ctx := c.Request.Context()

	var params AmbulanceSummaryQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.AllHistorySummary(ctx, params)
	c.JSON(res.Status, res)
}

// AmbulanceHistorySummary godoc
// @Summary Get ambulance history summary
// @Description Returns total service counts grouped by category for an ambulance.
// @Description Use the `period` query param to filter by time window.
// @Tags Ambulance History
// @Accept json
// @Produce json
// @Param id path string true "Ambulance ID"
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/{id}/history/summary [get]
func (h *handler) AmbulanceHistorySummary(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("id")

	var params AmbulanceSummaryQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.AmbulanceHistorySummary(ctx, ambulanceID, params)
	c.JSON(res.Status, res)
}

// CreateAmbulanceHistory godoc
// @Summary Create a new ambulance history record
// @Description Create a new ambulance history record with the provided details
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param ambulance_history body CreateAmbulanceHistoryRequest true "Ambulance history details"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-history [post]
func (h *handler) CreateAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	var payload CreateAmbulanceHistoryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request payload", nil, nil))
		return
	}
	res := h.service.CreateAmbulanceHistory(ctx, payload)
	c.JSON(res.Status, res)
}

// UpdateAmbulanceHistory godoc
// @Summary Update an existing ambulance history record
// @Description Update the details of an existing ambulance history record by ID
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Ambulance History ID"
// @Param ambulance_history body UpdateAmbulanceHistoryRequest true "Updated ambulance history details"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-history/{id} [put]
func (h *handler) UpdateAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceHistoryID := c.Param("id")
	var payload UpdateAmbulanceHistoryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request payload", nil, nil))
		return
	}
	res := h.service.UpdateAmbulanceHistory(ctx, ambulanceHistoryID, payload)
	c.JSON(res.Status, res)
}

// DeleteAmbulanceHistory godoc
// @Summary Delete an existing ambulance history record
// @Description Delete an existing ambulance history record by ID
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Ambulance History ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-history/{id} [delete]
func (h *handler) DeleteAmbulanceHistory(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceHistoryID := c.Param("id")
	res := h.service.DeleteAmbulanceHistory(ctx, ambulanceHistoryID)
	c.JSON(res.Status, res)
}

// DriverHistorySummary godoc
// @Summary Get driver's own ambulance history summary
// @Description Returns total service counts grouped by category for the authenticated driver.
// @Tags Ambulance History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} pkg.Response{data=SummaryResponse}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/driver/summary [get]
func (h *handler) DriverHistorySummary(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params AmbulanceSummaryQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.DriverHistorySummary(ctx, claims.AccountID, params)
	c.JSON(res.Status, res)
}

// DriverHistoryMonthlyTrend godoc
// @Summary Get driver's own ambulance history monthly trend
// @Description Retrieve aggregated monthly trend of ambulance history for a given year for the authenticated driver.
// @Tags Ambulance History
// @Security BearerAuth
// @Produce json
// @Param year query string false "Filter by year"
// @Success 200 {object} pkg.Response{data=HistoryMonthlyTrendRecord}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/history/driver/monthly-trend [get]
func (h *handler) DriverHistoryMonthlyTrend(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params MonthlyTrendQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.DriverHistoryMonthlyTrend(ctx, claims.AccountID, params)
	c.JSON(res.Status, res)
}
