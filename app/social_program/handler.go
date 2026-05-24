package social_program

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
	public := r.Group("/social-programs")
	public.Use(h.middleware.AuthOptional())
	{
		public.GET("", h.GetSocialProgramList)
		public.GET("/:slug", h.GetSocialProgramBySlug)
	}

	admin := r.Group("/admin/social-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleChairman, enum.RoleFinance))
	{
		admin.GET("", h.GetAdminSocialProgramList)
		admin.GET("/:id", h.GetAdminSocialProgramByID)
	}

	socialManager := r.Group("/admin/social-programs")
	socialManager.Use(h.middleware.RequireRoles(enum.RoleSocialManager))
	{
		socialManager.POST("", h.CreateSocialProgram)
		socialManager.PUT("/:id", h.UpdateSocialProgram)
		socialManager.DELETE("/:id", h.DeleteSocialProgram)
		socialManager.PATCH("/:id/complete", h.CompleteSocialProgram)
	}

	chairman := r.Group("/admin/social-programs")
	chairman.Use(h.middleware.RequireRoles(enum.RoleChairman))
	{
		chairman.PATCH("/:id/approve", h.ApproveSocialProgram)
		chairman.PATCH("/:id/reject", h.RejectSocialProgram)
	}
}

// GetSocialProgramList
//
// @Summary List Social Programs
// @Description Retrieve a list of social programs
// @Tags Social Programs
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param status query string false "Status filter"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/social-programs [get]
func (h *handler) GetSocialProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	accountID := ""
	role := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
			role = string(claims.ActiveRole)
		}
	}

	res := h.service.GetSocialProgramList(ctx, queryParams, false, accountID, role)
	c.JSON(res.Status, res)
}

// GetAdminSocialProgramList
//
// @Summary List Admin Social Programs
// @Description Retrieve a list of social programs for admin
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param status query string false "Status filter"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs [get]
func (h *handler) GetAdminSocialProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	accountID := ""
	role := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
			role = string(claims.ActiveRole)
		}
	}

	res := h.service.GetSocialProgramList(ctx, queryParams, true, accountID, role)
	c.JSON(res.Status, res)
}

// GetAdminSocialProgramByID
//
// @Summary Get Admin Social Program by ID
// @Description Get detailed information of a specific social program for admin
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id} [get]
func (h *handler) GetAdminSocialProgramByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	accountID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}

	res := h.service.GetSocialProgramByID(ctx, id, accountID)
	c.JSON(res.Status, res)
}

// GetSocialProgramByID
//
// @Summary Get Social Program by ID
// @Description Get detailed information of a specific social program
// @Tags Social Programs
// @Accept json
// @Produce json
// @Param slug path string true "Social Program Slug"
// @Success 200 {object} pkg.Response
// @Router /api/social-programs/{slug} [get]
func (h *handler) GetSocialProgramBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	accountID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}

	res := h.service.GetSocialProgramBySlug(ctx, slug, accountID)
	c.JSON(res.Status, res)
}

// CreateSocialProgram
//
// @Summary Create Social Program
// @Description Create a new social program entry
// @Tags Social Programs
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param payload formData SocialProgramRequest true "Social Program Data"
// @Param cover_image formData file true "Social Program Cover Image"
// @Success 201 {object} pkg.Response
// @Router /api/admin/social-programs [post]
func (h *handler) CreateSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()

	var req SocialProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateSocialProgram(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateSocialProgram
//
// @Summary Update Social Program
// @Description Update an existing social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Social Program ID"
// @Param payload formData SocialProgramRequest true "Social Program Data"
// @Param cover_image formData file false "Social Program Cover Image"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id} [put]
func (h *handler) UpdateSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req SocialProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.UpdateSocialProgram(ctx, id, req)
	c.JSON(res.Status, res)
}

// DeleteSocialProgram
//
// @Summary Delete Social Program
// @Description Delete a social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id} [delete]
func (h *handler) DeleteSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.DeleteSocialProgram(ctx, id)
	c.JSON(res.Status, res)
}

// CompleteSocialProgram
//
// @Summary Complete Social Program
// @Description Complete an active social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/complete [patch]
func (h *handler) CompleteSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.CompleteSocialProgram(ctx, id)
	c.JSON(res.Status, res)
}

// ApproveSocialProgram
//
// @Summary Approve Social Program
// @Description Approve a pending social program
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/activate [patch]
func (h *handler) ApproveSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.ActivateSocialProgram(ctx, id)
	c.JSON(res.Status, res)
}

// RejectSocialProgram
//
// @Summary Reject Social Program
// @Description Reject a pending social program with reason
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Param payload body RejectSocialProgramRequest true "Rejection Reason"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/reject [patch]
func (h *handler) RejectSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req RejectSocialProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.RejectSocialProgram(ctx, id, req)
	c.JSON(res.Status, res)
}
