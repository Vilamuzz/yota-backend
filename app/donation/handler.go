package donation

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service      Service
	s3Client     s3_pkg.Client
	middleware   middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, s3Client s3_pkg.Client, m middleware.AppMiddleware) {
	handler := &handler{
		service:      s,
		s3Client:     s3Client,
		middleware:   m,
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
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Donation Title"
// @Param description formData string true "Donation Description"
// @Param image formData file false "Donation Image File"
// @Param image_url formData string false "Donation Image URL"
// @Param category formData string true "Donation Category"
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

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil {
		// File uploaded, save to MinIO
		fileURL, err := h.s3Client.UploadFile(ctx, file, "donations")
		if err != nil {
			c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil))
			return
		}
		req.Image = fileURL
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
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Donation ID"
// @Param title formData string false "Donation Title"
// @Param description formData string false "Donation Description"
// @Param image formData file false "Donation Image File"
// @Param image_url formData string false "Donation Image URL"
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

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil {
		// File uploaded, save to MinIO
		fileURL, err := h.s3Client.UploadFile(ctx, file, "donations")
		if err != nil {
			c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil))
			return
		}
		req.Image = fileURL
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
