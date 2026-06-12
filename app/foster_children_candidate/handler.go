package foster_children_candidate

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
	user := r.Group("/foster-children/candidates")
	user.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		user.POST("", h.CreateFosterChildrenCandidate)
		user.GET("", h.GetMyFosterChildrenCandidateList)
		user.GET("/:id", h.GetMyFosterChildrenCandidateByID)
		user.DELETE("/:id", h.CancelFosterChildrenCandidate)
	}

	admin := r.Group("/admin/foster-children/candidates")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleChairman))
	{
		admin.GET("", h.GetFosterChildrenCandidateList)
		admin.GET("/:id", h.GetFosterChildrenCandidateByID)
		admin.PATCH("/:id/accept", h.AcceptFosterChildrenCandidate)
		admin.PATCH("/:id/reject", h.RejectFosterChildrenCandidate)
	}
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
// @Router /api/foster-children/candidates [post]
func (h *handler) CreateFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req CreateFosterChildrenCandidateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Payload request tidak valid", nil, nil))
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
// @Router /api/foster-children/candidates/{id} [delete]
func (h *handler) CancelFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")

	res := h.service.CancelFosterChildrenCandidate(ctx, claims.AccountID, id)
	c.JSON(res.Status, res)
}

// GetMyFosterChildrenCandidateByID
//
// @Summary Get My Foster Children Candidate By ID
// @Description Get a specific foster children candidate submitted by the user
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} pkg.Response
// @Router /api/foster-children/candidates/{id} [get]
func (h *handler) GetMyFosterChildrenCandidateByID(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	id := c.Param("id")

	res := h.service.GetMyFosterChildrenCandidateByID(ctx, claims.AccountID, id)
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
// @Router /api/foster-children/candidates [get]
func (h *handler) GetMyFosterChildrenCandidateList(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var queryParams FosterChildrenCandidateQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
		return
	}
	queryParams.AccountID = claims.AccountID

	res := h.service.GetMyFosterChildrenCandidateList(ctx, queryParams)
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
// @Param category query string false "Filter by category" Enums(yatim,piatu,yatim piatu,dhuafa)
// @Param gender query string false "Filter by gender" Enums(Laki-laki,Perempuan)
// @Param sortBy query string false "Sort by (e.g. name asc, created_at desc)"
// @Param search query string false "Search by name or submitter name"
// @Param page query int false "Page number"
// @Param limit query int false "Pagination limit"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates [get]
func (h *handler) GetFosterChildrenCandidateList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams FosterChildrenCandidateAdminQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
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

// AcceptFosterChildrenCandidate
//
// @Summary Accept Foster Children Candidate
// @Description Accept a foster children candidate. This is a two-step process: first by Social Manager (Koordinator Sosial) and then by Chairman (Ketua Yayasan).
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates/{id}/accept [patch]
func (h *handler) AcceptFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

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

	res := h.service.AcceptFosterChildrenCandidate(ctx, id, claims.ActiveRole)
	c.JSON(res.Status, res)
}

// RejectFosterChildrenCandidate
//
// @Summary Reject Foster Children Candidate
// @Description Reject a pending foster children candidate with a reason
// @Tags Foster Children Candidates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Candidate ID"
// @Param body body RejectFosterChildrenCandidateRequest true "Rejection Reason Request"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/candidates/{id}/reject [patch]
func (h *handler) RejectFosterChildrenCandidate(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req RejectFosterChildrenCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Payload request tidak valid", nil, nil))
		return
	}

	res := h.service.RejectFosterChildrenCandidate(ctx, id, req)
	c.JSON(res.Status, res)
}
