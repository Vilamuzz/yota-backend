package social_program

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/gin-gonic/gin"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
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
	public.GET("", h.GetSocialProgramList)
	public.GET("/:slug", h.GetSocialProgramBySlug)

	admin := r.Group("/admin/social-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager))
	{
		admin.GET("", h.GetAdminSocialProgramList)
		admin.POST("", h.CreateSocialProgram)
		admin.PUT("/:id", h.UpdateSocialProgram)
		admin.DELETE("/:id", h.DeleteSocialProgram)
	}

	chairman := r.Group("/admin/social-programs")
	chairman.Use(h.middleware.RequireRoles(enum.RoleChairman))
	{
		chairman.PUT("/:id/approve", h.ApproveSocialProgram)
		chairman.PUT("/:id/reject", h.RejectSocialProgram)
	}

	user := r.Group("/social-programs")
	user.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		user.POST("/:id/subscribe", h.SubscribeSocialProgram)
	}
}

func (h *handler) GetSocialProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter tidak valid", nil, nil))
		return
	}

	res := h.service.GetSocialProgramList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

func (h *handler) GetAdminSocialProgramList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams SocialProgramQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter tidak valid", nil, nil))
		return
	}

	res := h.service.GetSocialProgramList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

func (h *handler) GetSocialProgramBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	res := h.service.GetSocialProgramBySlug(ctx, slug)
	c.JSON(res.Status, res)
}

func (h *handler) CreateSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()

	var req SocialProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Request tidak valid", nil, nil))
		return
	}

	res := h.service.CreateSocialProgram(ctx, req)
	c.JSON(res.Status, res)
}

func (h *handler) UpdateSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req SocialProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Request tidak valid", nil, nil))
		return
	}

	res := h.service.UpdateSocialProgram(ctx, id, req)
	c.JSON(res.Status, res)
}

func (h *handler) DeleteSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.DeleteSocialProgram(ctx, id)
	c.JSON(res.Status, res)
}

func (h *handler) ApproveSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.ApproveSocialProgram(ctx, id)
	c.JSON(res.Status, res)
}

func (h *handler) RejectSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req SocialProgramRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Request tidak valid", nil, nil))
		return
	}

	res := h.service.RejectSocialProgram(ctx, id, req)
	c.JSON(res.Status, res)
}

func (h *handler) SubscribeSocialProgram(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	userData, exists := c.Get("user_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "User not authenticated", nil, nil))
		return
	}

	claims, ok := userData.(jwt_pkg.UserJWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Invalid user data", nil, nil))
		return
	}

	var req SocialProgramSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Request tidak valid", nil, nil))
		return
	}

	res := h.service.SubscribeSocialProgram(ctx, id, claims.AccountID, req)
	c.JSON(res.Status, res)
}