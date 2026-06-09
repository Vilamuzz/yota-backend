package ambulance_service_request

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
	public := r.Group("/ambulances/requests")
	public.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		public.POST("", h.CreateAmbulanceServiceRequest)
		public.GET("", h.ListMyAmbulanceServiceRequests)
		public.GET("/:id", h.GetMyAmbulanceServiceRequestByID)
		public.PATCH("/:id/cancel", h.CancelAmbulanceServiceRequest)
	}

	ambulanceManager := r.Group("/admin/ambulances/requests")
	ambulanceManager.Use(h.middleware.RequireRoles(enum.RoleAmbulanceManager))
	{
		ambulanceManager.GET("", h.ListAmbulanceServiceRequests)
		ambulanceManager.GET("/:id", h.GetAmbulanceServiceRequestByID)
		ambulanceManager.PATCH("/:id/accept", h.AcceptAmbulanceServiceRequest)
		ambulanceManager.PATCH("/:id/reject", h.RejectAmbulanceServiceRequest)
	}

	ambulanceDriver := r.Group("/admin/ambulances/requests/assigned")
	ambulanceDriver.Use(h.middleware.RequireRoles(enum.RoleAmbulanceDriver))
	{
		ambulanceDriver.GET("/:ambulanceId", h.ListAssignedAmbulanceServiceRequests)
		ambulanceDriver.GET("/:ambulanceId/detail/:id", h.GetAssignedAmbulanceServiceRequestByID)
		ambulanceDriver.PATCH("/:ambulanceId/start/:id", h.StartAmbulanceServiceRequest)
		ambulanceDriver.PATCH("/:ambulanceId/complete/:id", h.CompleteAmbulanceServiceRequest)
	}
}

// ListMyAmbulanceRequests godoc
// @Summary List my ambulance requests
// @Description Get a list of ambulance requests created by the authenticated user with pagination
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param search query string false "Search by applicant name"
// @Param sortBy query string false "Sort by (e.g. applicant_name asc, created_at desc)"
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/requests/me [get]
func (h *handler) ListMyAmbulanceServiceRequests(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	var queryParams AmbulanceServiceRequestQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListMyAmbulanceServiceRequests(ctx, claims.AccountID, queryParams)
	c.JSON(res.Status, res)
}

// GetMyAmbulanceServiceRequestByID
//
// @Summary Get My Ambulance Service Request By ID
// @Description Get a specific ambulance service request submitted by the user
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} pkg.Response
// @Router /ambulances/requests/{id} [get]
func (h *handler) GetMyAmbulanceServiceRequestByID(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")

	res := h.service.GetMyAmbulanceServiceRequestByID(ctx, claims.AccountID, id)
	c.JSON(res.Status, res)
}

// ListAmbulanceRequests godoc
// @Summary List ambulance requests
// @Description Get a list of ambulance requests with pagination
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param account_id query string false "Filter by account id"
// @Param sortBy query string false "Sort by (e.g. applicant_name asc, created_at desc)"
// @Param search query string false "Search by applicant name"
// @Param page query int false "Page number"
// @Param limit query int false "Pagination limit"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/requests [get]
func (h *handler) ListAmbulanceServiceRequests(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams AmbulanceServiceRequestAdminQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListAmbulanceServiceRequest(ctx, queryParams)
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

// ListAssignedAmbulanceServiceRequests godoc
// @Summary List assigned ambulance requests (driver)
// @Description Get a list of ambulance requests assigned to the authenticated ambulance driver
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param search query string false "Search by applicant name"
// @Param sortBy query string false "Sort by (e.g. applicant_name asc, created_at desc)"
// @Param limit query int false "Number of items to return"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/requests/assigned [get]
func (h *handler) ListAssignedAmbulanceServiceRequests(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	ambulanceID := c.Param("ambulanceId")
	var queryParams AmbulanceServiceRequestQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListAssignedAmbulanceServiceRequests(ctx, claims.AccountID, ambulanceID, queryParams)
	c.JSON(res.Status, res)
}

// GetAssignedAmbulanceServiceRequestByID godoc
// @Summary Get assigned ambulance request by ID (driver)
// @Description Get the details of an ambulance request assigned to the authenticated driver
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance Request ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 403 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /admin/ambulances/requests/assigned/{ambulanceId}/detail/{id} [get]
func (h *handler) GetAssignedAmbulanceServiceRequestByID(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("ambulanceId")
	id := c.Param("id")

	res := h.service.GetAssignedAmbulanceServiceRequestByID(ctx, ambulanceID, id)
	c.JSON(res.Status, res)
}

// CreateAmbulanceServiceRequest godoc
// @Summary Create a new ambulance service request
// @Description Create a new ambulance service request with the provided details
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param payload formData CreateAmbulanceServiceRequest true "Ambulance service request details"
// @Param submitterIdCard formData file true "Submitter ID Card Image"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /ambulances/requests [post]
func (h *handler) CreateAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	var payload CreateAmbulanceServiceRequest
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request body", nil, nil))
		return
	}
	payload.AccountID = claims.AccountID
	res := h.service.CreateAmbulanceServiceRequest(ctx, payload)
	c.JSON(res.Status, res)
}

// AcceptAmbulanceService
//
// @Summary Accept Ambulance Service Request
// @Description Accept an ambulance service request and assign it to an ambulance
// @Tags Ambulance Service Requests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ambulance Request ID"
// @Param payload body AcceptAmbulanceServiceRequestPayload true "Ambulance ID to assign"
// @Success 200 {object} pkg.Response
// @Router /admin/ambulances/requests/{id}/accept [patch]
func (h *handler) AcceptAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var payload AcceptAmbulanceServiceRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request payload", nil, nil))
		return
	}

	userData, exists := c.Get("user_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Tidak terautorisasi", nil, nil))
		return
	}

	claims, ok := userData.(jwt_pkg.UserJWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Gagal memproses data user", nil, nil))
		return
	}

	res := h.service.AcceptAmbulanceServiceRequest(ctx, id, claims.ActiveRole, payload)
	c.JSON(res.Status, res)
}

// RejectAmbulanceService
//
// @Summary Reject Foster Children Candidate
// @Description Reject a pending foster children candidate with a reason
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Candidate ID"
// @Param body body RejectAmbulanceServiceRequest true "Rejection Reason Request"
// @Success 200 {object} pkg.Response
// @Router /ambulances/requests/{id}/reject [patch]
func (h *handler) RejectAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req RejectAmbulanceServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Payload request tidak valid", nil, nil))
		return
	}

	res := h.service.RejectAmbulanceServiceRequest(ctx, id, req)
	c.JSON(res.Status, res)
}

func (h *handler) CancelAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")
	res := h.service.CancelAmbulanceServiceRequest(ctx, claims.AccountID, id)
	c.JSON(res.Status, res)
}

func (h *handler) StartAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	ambulanceID := c.Param("ambulanceId")
	id := c.Param("id")

	res := h.service.StartAmbulanceServiceRequest(ctx, claims.AccountID, ambulanceID, id)
	c.JSON(res.Status, res)
}

func (h *handler) CompleteAmbulanceServiceRequest(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	ambulanceID := c.Param("ambulanceId")
	id := c.Param("id")

	res := h.service.CompleteAmbulanceServiceRequest(ctx, claims.AccountID, ambulanceID, id)
	c.JSON(res.Status, res)
}
