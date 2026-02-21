package donation

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
	handler := &handler{
		service:    s,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	// Public routes
	public := r.Group("/public/donations")
	public.GET("", h.ListPublishedDonations)
	public.GET("/:id", h.GetPublishedDonation)

	// Protected routes
	protected := r.Group("/donations")
	protected.Use(h.middleware.RequireRoles(enum.RoleSuperadmin, enum.RoleFinance))
	{
		protected.GET("", h.ListDonations)
		protected.GET("/:id", h.GetDonationByID)
		protected.POST("/", h.CreateDonation)
		protected.PUT("/:id", h.UpdateDonation)
		protected.DELETE("/:id", h.DeleteDonation)
	}
}

// ListPublishedDonations
//
// @Summary List Published Donations
// @Description Retrieve a list of published (active) donations
// @Tags Donations
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/public/donations/ [get]
func (h *handler) ListPublishedDonations(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.ListPublished(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetPublishedDonation
//
// @Summary Get Published Donation by ID
// @Description Get detailed information of a specific published donation
// @Tags Donations
// @Accept json
// @Produce json
// @Param id path string true "Donation ID"
// @Success 200 {object} pkg.Response
// @Router /api/public/donations/{id} [get]
func (h *handler) GetPublishedDonation(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.GetPublishedByID(ctx, donationID)
	c.JSON(res.Status, res)
}

// ListDonations
//
// @Summary Get All Donations
// @Description Retrieve a list of all donations with cursor-based pagination and optional filters
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (education, health, environment)"
// @Param status query string false "Filter by status (active, inactive, completed)"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/donations/ [get]
func (h *handler) ListDonations(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.List(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetDonationByID
//
// @Summary Get Donation by ID
// @Description Get detailed information of a specific donation
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation ID"
// @Success 200 {object} pkg.Response
// @Router /api/donations/{id} [get]
func (h *handler) GetDonationByID(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.GetByID(ctx, donationID)
	c.JSON(res.Status, res)
}

// CreateDonation
//
// @Summary Create Donation
// @Description Create a new donation entry (requires authentication and proper role)
// @Tags Donations
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Donation Title"
// @Param description formData string true "Donation Description"
// @Param image formData file false "Donation Image File"
// @Param category formData string true "Donation Category"
// @Param status formData string false "Donation Status"
// @Param fund_target formData number true "Fund Target"
// @Param date_end formData string true "End Date (RFC3339)"
// @Success 201 {object} pkg.Response
// @Router /api/donations/ [post]
func (h *handler) CreateDonation(c *gin.Context) {
	ctx := c.Request.Context()

	var req DonationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateDonation(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateDonation
//
// @Summary Update Donation
// @Description Update an existing donation (requires authentication and proper role)
// @Tags Donations
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Donation ID"
// @Param title formData string false "Donation Title"
// @Param description formData string false "Donation Description"
// @Param image formData file false "Donation Image File"
// @Param category formData string false "Donation Category"
// @Param fund_target formData number false "Fund Target"
// @Param status formData string false "Status"
// @Param date_end formData string false "End Date (RFC3339)"
// @Success 200 {object} pkg.Response
// @Router /api/donations/{id} [put]
func (h *handler) UpdateDonation(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	var req UpdateDonationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.UpdateDonation(ctx, donationID, req)
	c.JSON(res.Status, res)
}

// DeleteDonation
//
// @Summary Delete Donation
// @Description Delete a donation (requires authentication and proper role)
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation ID"
// @Success 200 {object} pkg.Response
// @Router /api/donations/{id} [delete]
func (h *handler) DeleteDonation(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.DeleteDonation(ctx, donationID)
	c.JSON(res.Status, res)
}
