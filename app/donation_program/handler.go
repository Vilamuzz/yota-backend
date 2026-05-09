package donation_program

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
	public := r.Group("/donation-programs")
	public.GET("", h.GetDonationProgramList)
	public.GET("/:slug", h.GetDonationProgramBySlug)

	// Admin routes
	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("", h.GetAdminDonationProgramList)
		admin.GET("/:id", h.GetDonationProgramByID)
		admin.POST("", h.CreateDonationProgram)
		admin.PUT("/:id", h.UpdateDonationProgram)
		admin.DELETE("/:id", h.DeleteDonationProgram)
		admin.PATCH("/:id/active", h.UpdateActiveDonationProgram)
		admin.PATCH("/:id/archive", h.UpdateArchiveDonationProgram)
	}
}

// GetDonationProgramList
//
// @Summary List Donation Programs
// @Description Retrieve a list of donation programs
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param category query string false "Filter by category"
// @Param status query string false "Status filter"
// @Param limit query int false "Pagination limit"
// @Param nextCursor query string false "Pagination cursor (next page)"
// @Param prevCursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/donation-programs [get]
func (h *handler) GetDonationProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetDonationProgramList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

// GetDonationProgramBySlug
//
// @Summary Get Donation Program by Slug
// @Description Get detailed information of a specific donation program
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Success 200 {object} pkg.Response
// @Router /api/donation-programs/{slug} [get]
func (h *handler) GetDonationProgramBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	donationSlug := c.Param("slug")

	res := h.service.GetDonationProgramBySlug(ctx, donationSlug)
	c.JSON(res.Status, res)
}

// GetAdminDonationProgramList
//
// @Summary Get All Donation Programs
// @Description Retrieve a list of all donation programs with cursor-based pagination and optional filters
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param category query string false "Filter by category"
// @Param status query string false "Status filter"
// @Param limit query int false "Pagination limit"
// @Param nextCursor query string false "Pagination cursor (next page)"
// @Param prevCursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs [get]
func (h *handler) GetAdminDonationProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetDonationProgramList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

// GetDonationProgramByID
//
// @Summary Get Donation Program by ID
// @Description Get detailed information of a specific donation program
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id} [get]
func (h *handler) GetDonationProgramByID(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.GetDonationProgramByID(ctx, donationID)
	c.JSON(res.Status, res)
}

// CreateDonationProgram
//
// @Summary Create Donation Program
// @Description Create a new donation program entry (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Donation Program Title"
// @Param coverImage formData file true "Cover Image"
// @Param category formData string true "Category"
// @Param description formData string true "Description"
// @Param fundTarget formData number true "Fund Target"
// @Param status formData string false "Status"
// @Param startDate formData string true "Start Date (YYYY-MM-DD)"
// @Param endDate formData string true "End Date (YYYY-MM-DD)"
// @Success 201 {object} pkg.Response
// @Router /api/admin/donation-programs [post]
func (h *handler) CreateDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()

	var req DonationProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.CreateDonationProgram(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateDonationProgram
//
// @Summary Update Donation Program
// @Description Update an existing donation program (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Donation Program ID"
// @Param title formData string false "Donation Program Title"
// @Param coverImage formData file false "Cover Image"
// @Param category formData string false "Category"
// @Param description formData string false "Description"
// @Param fundTarget formData number false "Fund Target"
// @Param status formData string false "Status"
// @Param startDate formData string false "Start Date (YYYY-MM-DD)"
// @Param endDate formData string false "End Date (YYYY-MM-DD)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id} [put]
func (h *handler) UpdateDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	var req DonationProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.UpdateDonationProgram(ctx, donationID, req)
	c.JSON(res.Status, res)
}

// DeleteDonationProgram
//
// @Summary Delete Donation Program
// @Description Delete a donation program (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id} [delete]
func (h *handler) DeleteDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.DeleteDonationProgram(ctx, donationID)
	c.JSON(res.Status, res)
}

// UpdateActiveDonationProgram
//
// @Summary Update Active Donation Program
// @Description Update an existing donation program to active (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/active [patch]
func (h *handler) UpdateActiveDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.UpdateActiveDonationProgram(ctx, donationID)
	c.JSON(res.Status, res)
}

// UpdateArchiveDonationProgram
//
// @Summary Update Archive Donation Program
// @Description Update an existing donation program to archived (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/archive [patch]
func (h *handler) UpdateArchiveDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	res := h.service.UpdateArchivedDonationProgram(ctx, donationID)
	c.JSON(res.Status, res)
}
