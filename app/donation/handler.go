package donation

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
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
	api := r.Group("/donations")

	// Public routes
	api.GET("/", h.GetAllDonations)
	api.GET("/:id", h.GetDonationByID)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(h.middleware.AuthRequired())
	{
		protected.POST("/", h.middleware.RequireRoles(string(user.RoleSuperadmin), string(user.RoleChairman), string(user.RoleSocialManager)), h.CreateDonation)
		protected.PUT("/:id", h.middleware.RequireRoles(string(user.RoleSuperadmin), string(user.RoleChairman), string(user.RoleSocialManager)), h.UpdateDonation)
		protected.DELETE("/:id", h.middleware.RequireRoles(string(user.RoleSuperadmin), string(user.RoleChairman)), h.DeleteDonation)
	}
}

// GetAllDonations
//
// @Summary Get All Donations
// @Description Retrieve a list of all donations with cursor-based pagination and optional filters
// @Tags Donations
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (education, health, environment)"
// @Param status query string false "Filter by status (active, inactive, completed)"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/donations/ [get]
func (h *handler) GetAllDonations(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.FetchAllDonations(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetDonationByID
//
// @Summary Get Donation by ID
// @Description Get detailed information of a specific donation
// @Tags Donations
// @Accept json
// @Produce json
// @Param id path string true "Donation ID"
// @Success 200 {object} pkg.Response
// @Router /api/donations/{id} [get]
func (h *handler) GetDonationByID(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.FetchDonationByID(ctx, donationID)
	c.JSON(res.Status, res)
}

// CreateDonation
//
// @Summary Create Donation
// @Description Create a new donation entry (requires authentication and proper role)
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body DonationRequest true "Create Donation"
// @Success 201 {object} pkg.Response
// @Router /api/donations/ [post]
func (h *handler) CreateDonation(c *gin.Context) {
	ctx := c.Request.Context()

	var req DonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.Create(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateDonation
//
// @Summary Update Donation
// @Description Update an existing donation (requires authentication and proper role)
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation ID"
// @Param payload body UpdateDonationRequest true "Update Donation"
// @Success 200 {object} pkg.Response
// @Router /api/donations/{id} [put]
func (h *handler) UpdateDonation(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	var req UpdateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.Update(ctx, donationID, req)
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

	res := h.service.Delete(ctx, donationID)
	c.JSON(res.Status, res)
}
