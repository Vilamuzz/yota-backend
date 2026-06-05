package foster_children

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
	// Public routes
	public := r.Group("/foster-children")
	public.GET("", h.GetFosterChildrenList)
	public.GET("/:slug", h.GetFosterChildrenBySlug)

	adminFosterChildren := r.Group("/admin/foster-children")
	adminFosterChildren.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleFinance))
	{
		adminFosterChildren.GET("", h.GetAdminFosterChildrenList)
		adminFosterChildren.GET("/:id", h.GetAdminFosterChildrenByID)
	}

	socialManagerOnly := r.Group("/admin/foster-children")
	socialManagerOnly.Use(h.middleware.RequireRoles(enum.RoleSocialManager))
	{
		socialManagerOnly.POST("", h.CreateFosterChildren)
		socialManagerOnly.PUT("/:id", h.UpdateFosterChildren)
		socialManagerOnly.DELETE("/:id", h.DeleteFosterChildren)
	}

}

// GetFosterChildrenList
//
// @Summary List Foster Children
// @Description Retrieve a paginated list of foster children with optional filters and sorting
// @Tags Foster Children
// @Accept json
// @Produce json
// @Param search query string false "Search by name"
// @Param category query string false "Filter by category"
// @Param gender query string false "Filter by gender (male, female)"
// @Param isGraduated query boolean false "Filter by graduation status"
// @Param educationLevel query int false "Filter by education level (1-12)"
// @Param sortBy query string false "Sort field and direction, e.g. 'name asc', 'education_level desc', 'birth_date asc', 'created_at desc'"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/foster-children [get]
func (h *handler) GetFosterChildrenList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams FosterChildrenQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
		return
	}

	res := h.service.GetFosterChildrenList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

func (h *handler) GetAdminFosterChildrenList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams FosterChildrenQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Parameter query tidak valid", nil, nil))
		return
	}

	res := h.service.GetFosterChildrenList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

// GetFosterChildrenByID
//
// @Summary Get Foster Children by ID
// @Description Get detailed information of a specific foster child
// @Tags Foster Children
// @Accept json
// @Produce json
// @Param slug path string true "Foster Children Slug"
// @Success 200 {object} pkg.Response
// @Router /api/foster-children/{slug} [get]
func (h *handler) GetFosterChildrenBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	res := h.service.GetFosterChildrenBySlug(ctx, slug)
	c.JSON(res.Status, res)
}

func (h *handler) GetAdminFosterChildrenByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Body request tidak valid", nil, nil))
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
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Body request tidak valid", nil, nil))
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

