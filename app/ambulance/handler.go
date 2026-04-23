package ambulance

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
	h := &handler{
		service:    s,
		middleware: m,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	public := r.Group("/ambulances")
	public.GET("", h.ListAmbulances)
	public.GET("/:id", h.GetAmbulanceByID)

	// Protected routes
	protected := r.Group("/ambulances")
	protected.Use(h.middleware.RequireRoles(enum.RoleAmbulanceManager))
	{
		protected.POST("/", h.CreateAmbulance)
		protected.PUT("/:id", h.UpdateAmbulance)
		protected.DELETE("/:id", h.DeleteAmbulance)
	}
}

// ListAmbulances godoc
// @Summary List ambulances
// @Description Get a list of ambulances with pagination
// @Tags Ambulances
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances [get]
func (h *handler) ListAmbulances(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams AmbulanceQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListAmbulance(ctx, queryParams)
	c.JSON(res.Status, res)
}

func (h *handler) GetAmbulanceByID(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("id")
	res := h.service.FindAmbulanceById(ctx, ambulanceID)
	c.JSON(res.Status, res)
}

// CreateAmbulance godoc
// @Summary Create a new ambulance
// @Description Create a new ambulance with the provided details
// @Tags Ambulances
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param ambulance formData CreateAmbulanceRequest true "Ambulance details"
// @Param image formData file false "Ambulance Image"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances [post]
func (h *handler) CreateAmbulance(c *gin.Context) {
	ctx := c.Request.Context()
	var req CreateAmbulanceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.CreateAmbulance(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateAmbulance godoc
// @Summary Update an existing ambulance
// @Description Update the details of an existing ambulance by ID
// @Tags Ambulances
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param id path string true "Ambulance ID"
// @Param ambulance formData UpdateAmbulanceRequest true "Updated ambulance details"
// @Param image formData file false "Ambulance Image"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/{id} [put]
func (h *handler) UpdateAmbulance(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("id")

	var req UpdateAmbulanceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.UpdateAmbulance(ctx, ambulanceID, req)
	c.JSON(res.Status, res)
}

// DeleteAmbulance godoc
// @Summary Delete an existing ambulance
// @Description Delete an existing ambulance by ID
// @Tags Ambulances
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/{id} [delete]
func (h *handler) DeleteAmbulance(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("id")
	res := h.service.DeleteAmbulance(ctx, ambulanceID)
	c.JSON(res.Status, res)
}
