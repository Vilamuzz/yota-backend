package ambulance_service_request

import (
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
	public := r.Group("/ambulance-service-requests")
	public.Use(h.middleware.AuthRequired())
	{
		public.POST("", h.CreateAmbulanceServiceRequest)
		public.GET("/me", h.ListMyAmbulanceServiceRequests)
		public.GET("/:id", h.GetAmbulanceServiceRequestByID)
	}

	protected := r.Group("/ambulance-service-requests")
	protected.Use(h.middleware.RequireRoles(enum.RoleAmbulanceDriver, enum.RoleAmbulanceManager))
	{
		protected.GET("/", h.ListAmbulanceServiceRequests)
		protected.PUT("/:id", h.UpdateAmbulanceServiceRequest)
	}
}

// ListMyAmbulanceRequests godoc
// @Summary List my ambulance requests
// @Description Get a list of ambulance requests created by the authenticated user with pagination
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-service-requests/me [get]
func (h *handler) ListMyAmbulanceServiceRequests(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.MustGet("accountID").(string)
	var queryParams AmbulanceServiceRequestQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}
	queryParams.AccountID = accountID
	res := h.service.ListAmbulanceServiceRequest(ctx, queryParams)
	c.JSON(res.Status, res)
}

// ListAmbulanceRequests godoc
// @Summary List ambulance requests
// @Description Get a list of ambulance requests with pagination
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-service-requests [get]
func (h *handler) ListAmbulanceServiceRequests(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams AmbulanceServiceRequestQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListAmbulanceServiceRequest(ctx, queryParams)
	c.JSON(res.Status, res)
}

func (h *handler) CreateAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.MustGet("accountID").(string)
	var payload CreateAmbulanceServiceRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request body", nil, nil))
		return
	}
	payload.AccountID = accountID
	res := h.service.CreateAmbulanceServiceRequest(ctx, payload)
	c.JSON(res.Status, res)
}

// GetAmbulanceRequestByID godoc
// @Summary Get ambulance request by ID
// @Description Get the details of an ambulance request by its ID
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance Request ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-service-requests/{id} [get]
func (h *handler) GetAmbulanceServiceRequestByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.GetAmbulanceServiceRequestByID(ctx, id)
	c.JSON(res.Status, res)
}

// UpdateAmbulanceRequest godoc
// @Summary Update ambulance request
// @Description Update the details of an ambulance request by its ID
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance Request ID"
// @Param ambulance_service_request body UpdateAmbulanceServiceRequest true "Ambulance request details to update"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulance-service-requests/{id} [put]
func (h *handler) UpdateAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var payload UpdateAmbulanceServiceRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request body", nil, nil))
		return
	}
	res := h.service.UpdateAmbulanceServiceRequest(ctx, id, payload)
	c.JSON(res.Status, res)
}
