package foster_children

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
	// Public routes
	public := r.Group("/foster-children")
	public.GET("", h.GetFosterChildrenList)
	public.GET("/:id", h.GetFosterChildrenByID)

	user := r.Group("/foster-children/candidates")
	user.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		user.POST("/submit", h.CreateFosterChildrenCandidate)
	}

	me := r.Group("/me/foster-children")
	me.Use(h.middleware.AuthRequired())
	{
		me.GET("/candidates", h.GetMyFosterChildrenCandidateList)
		me.DELETE("/candidates/:id", h.CancelFosterChildrenCandidate)
	}

	admin := r.Group("/admin")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleChairman))
	{
		fosterChildren := admin.Group("/foster-children")
		fosterChildren.POST("", h.CreateFosterChildren)
		fosterChildren.PUT("/:id", h.UpdateFosterChildren)
		fosterChildren.DELETE("/:id", h.DeleteFosterChildren)

		fosterCandidate := fosterChildren.Group("/candidates")
		fosterCandidate.GET("", h.GetFosterChildrenCandidateList)
		fosterCandidate.GET("/:id", h.GetFosterChildrenCandidateByID)
		fosterCandidate.PATCH("/:id", h.UpdateFosterChildrenCandidateStatus)
	}
}

// GetFosterChildrenList
//
// @Summary List Foster Children
// @Description Retrieve a list of foster children with cursor-based pagination and optional filters
// @Tags Foster Children
// @Accept json
// @Produce json
// @Param search query string false "Search by name"
// @Param category query string false "Filter by category"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/foster-children [get]
func (h *handler) GetFosterChildrenList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams FosterChildrenQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetFosterChildrenList(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetFosterChildrenByID
//
// @Summary Get Foster Children by ID
// @Description Get detailed information of a specific foster child
// @Tags Foster Children
// @Accept json
// @Produce json
// @Param id path string true "Foster Children ID"
// @Success 200 {object} pkg.Response
// @Router /api/foster-children/{id} [get]
func (h *handler) GetFosterChildrenByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetFosterChildrenByID(ctx, id)
	c.JSON(res.Status, res)
}

// CreateFosterChildren
//
// @Summary Create Foster Children
// @Description Create a new foster child entry (requires authentication and social_manager role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param payload formData CreateFosterChildrenRequest true "Foster Children Data"
// @Param profile_picture formData file true "Profile Picture"
// @Param family_card formData file true "Family Card"
// @Param sktm formData file true "SKTM"
// @Param achievements formData file false "Achievements"
// @Success 201 {object} pkg.Response
// @Router /api/admin/foster-children [post]
func (h *handler) CreateFosterChildren(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateFosterChildrenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateFosterChildren(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateFosterChildren
//
// @Summary Update Foster Children
// @Description Update an existing foster child entry (requires authentication and social_manager role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Foster Children ID"
// @Param payload formData UpdateFosterChildrenRequest true "Foster Children Data"
// @Param profile_picture formData file false "Profile Picture"
// @Param family_card formData file false "Family Card"
// @Param sktm formData file false "SKTM"
// @Param achievements formData file false "Achievements"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/{id} [put]
func (h *handler) UpdateFosterChildren(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req UpdateFosterChildrenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.UpdateFosterChildren(ctx, id, req)
	c.JSON(res.Status, res)
}

// DeleteFosterChildren
//
// @Summary Delete Foster Children
// @Description Delete a foster child entry (requires authentication and social_manager role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Foster Children ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/{id} [delete]
func (h *handler) DeleteFosterChildren(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.DeleteFosterChildren(ctx, id)
	c.JSON(res.Status, res)
}

// CreateFosterChildrenCandidate
//
// @Summary Create Foster Children Candidate
// @Description Submit a request to propose a new foster child candidate
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param payload formData CreateFosterChildrenCandidateRequest true "Candidate Request"
// @Param profile_picture formData file true "Profile Picture"
// @Param family_card formData file true "Family Card"
// @Param sktm formData file true "SKTM"
// @Param submitter_id_card formData file true "Submitter ID Card"
// @Success 201 {object} pkg.Response
// @Router /api/foster-children/candidates/submit [post]
func (h *handler) CreateFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req CreateFosterChildrenCandidateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request payload", nil, nil))
		return
	}

	res := h.service.CreateFosterChildrenCandidate(ctx, claims.AccountID, req)
	c.JSON(res.Status, res)
}

// CancelFosterChildrenCandidate
//
// @Summary Cancel Foster Children Candidate
// @Description Cancel a pending foster children candidate
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/foster-children/candidates/{id} [delete]
func (h *handler) CancelFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")

	res := h.service.CancelFosterChildrenCandidate(ctx, claims.AccountID, id)
	c.JSON(res.Status, res)
}

// GetMyFosterChildrenCandidateList
//
// @Summary List My Foster Children Candidates
// @Description Get a list of the authenticated user's foster children candidates
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/me/foster-children/candidates [get]
func (h *handler) GetMyFosterChildrenCandidateList(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var queryParams FosterChildrenCandidateQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	queryParams.AccountID = claims.AccountID

	res := h.service.GetFosterChildrenCandidateList(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetFosterChildrenCandidateList
//
// @Summary List All Foster Children Candidates
// @Description Get a list of all foster children candidates (admin only)
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param account_id query string false "Filter by account id"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates [get]
func (h *handler) GetFosterChildrenCandidateList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams FosterChildrenCandidateQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetFosterChildrenCandidateList(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetFosterChildrenCandidateByID
//
// @Summary Get Foster Children Candidate By ID
// @Description Get a specific foster children candidate
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates/{id} [get]
func (h *handler) GetFosterChildrenCandidateByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetFosterChildrenCandidateByID(ctx, id)
	c.JSON(res.Status, res)
}

// UpdateFosterChildrenCandidateStatus
//
// @Summary Update Foster Children Candidate Status
// @Description Accept or reject a foster children candidate
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Candidate ID"
// @Param body body UpdateFosterChildrenCandidateStatusRequest true "Status Update Request"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates/{id} [patch]
func (h *handler) UpdateFosterChildrenCandidateStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req UpdateFosterChildrenCandidateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request payload", nil, nil))
		return
	}

	res := h.service.UpdateFosterChildrenCandidateStatus(ctx, id, req)
	c.JSON(res.Status, res)
}
