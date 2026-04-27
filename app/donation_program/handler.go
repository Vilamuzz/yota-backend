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
	public.GET("", h.GetPublishedDonationProgramList)
	public.GET("/:slug", h.GetPublishedDonationProgramBySlug)

	// Admin routes
	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("", h.GetDonationProgramList)
		admin.GET("/:id", h.GetDonationProgramByID)
		admin.POST("", h.CreateDonationProgram)
		admin.PUT("/:id", h.UpdateDonationProgram)
		admin.DELETE("/:id", h.DeleteDonationProgram)
	}
}

// GetPublishedDonationProgramList
//
// @Summary List Published Donation Programs
// @Description Retrieve a list of published (active) donation programs
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param limit query int false "Pagination limit"
// @Param nextCursor query string false "Pagination cursor (next page)"
// @Param prevCursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/donation-programs [get]
func (h *handler) GetPublishedDonationProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams DonationProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetDonationProgramList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

// GetPublishedDonationProgramBySlug
//
// @Summary Get Published Donation Program by Slug
// @Description Get detailed information of a specific published donation program
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Success 200 {object} pkg.Response
// @Router /api/donation-programs/{slug} [get]
func (h *handler) GetPublishedDonationProgramBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	donationSlug := c.Param("slug")

	res := h.service.GetPublishedDonationProgramBySlug(ctx, donationSlug)
	c.JSON(res.Status, res)
}

// GetDonationProgramList
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
func (h *handler) GetDonationProgramList(c *gin.Context) {
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
// @Param payload formData DonationProgramRequest true "Donation Program Data"
// @Param coverImage formData file true "Donation Program Cover Image"
// @Success 201 {object} pkg.Response
// @Router /api/admin/donation-programs [post]
func (h *handler) CreateDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()

	var req DonationProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
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
// @Param payload formData DonationProgramRequest true "Donation Program Data"
// @Param coverImage formData file false "Donation Program Cover Image"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id} [put]
func (h *handler) UpdateDonationProgram(c *gin.Context) {
	ctx := c.Request.Context()
	donationID := c.Param("id")

	var req DonationProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
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
