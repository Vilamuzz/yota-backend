package news

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
	public := r.Group("/news")
	public.GET("", h.ListPublishedNews)
	public.GET("/:slug", h.GetNewsBySlug)

	admin := r.Group("/admin/news")
	admin.Use(h.middleware.RequireRoles(enum.RolePublicationManager, enum.RoleSuperadmin))
	{
		admin.GET("", h.ListNews)
		admin.GET("/:id", h.GetNews)
		admin.POST("", h.CreateNews)
		admin.PUT("/:id", h.UpdateNews)
		admin.DELETE("/:id", h.DeleteNews)
		admin.PATCH("/:id/published", h.UpdatePublishedNews)
		admin.PATCH("/:id/archived", h.UpdateArchivedNews)
	}
}

// ListPublishedNews
//
// @Summary List Published News
// @Description Retrieve a list of published news with cursor-based pagination and optional filters
// @Tags News
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/news/ [get]
func (h *handler) ListPublishedNews(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams NewsQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetNewsList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

// GetNewsBySlug
//
// @Summary Get Published News
// @Description Get detailed information of a specific published news article
// @Tags News
// @Accept json
// @Produce json
// @Param slug path string true "News Slug"
// @Success 200 {object} pkg.Response
// @Router /api/news/{slug} [get]
func (h *handler) GetNewsBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	res := h.service.GetNewsBySlug(ctx, slug)
	c.JSON(res.Status, res)
}

// ListNews
//
// @Summary List All News (Protected)
// @Description Retrieve a list of all news (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/ [get]
func (h *handler) ListNews(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams NewsQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetNewsList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

// GetNews
//
// @Summary Get News (Protected)
// @Description Get detailed information of a specific news article (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id} [get]
func (h *handler) GetNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.GetNewsByID(ctx, newsID)
	c.JSON(res.Status, res)
}

// CreateNews
//
// @Summary Create News
// @Description Create a new news entry (requires authentication and proper role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "News Title"
// @Param category formData string true "News Category"
// @Param content formData string true "News Content"
// @Param status formData string false "News Status"
// @Param coverImage formData file true "Cover Image"
// @Param metadata formData string false "Media metadata JSON array"
// @Param files formData file false "Additional Media Files"
// @Success 201 {object} pkg.Response
// @Router /api/admin/news [post]
func (h *handler) CreateNews(c *gin.Context) {
	ctx := c.Request.Context()

	var req NewsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}
	resp := h.service.CreateNews(ctx, req)
	c.JSON(resp.Status, resp)
}

// UpdateNews
//
// @Summary Update News
// @Description Update an existing news entry (requires authentication and proper role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "News ID"
// @Param title formData string false "News Title"
// @Param category formData string false "News Category"
// @Param content formData string false "News Content"
// @Param status formData string false "News Status"
// @Param coverImage formData file false "Cover Image"
// @Param metadata formData string false "Media metadata JSON array"
// @Param files formData file false "Additional Media Files"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id} [put]
func (h *handler) UpdateNews(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req NewsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}
	resp := h.service.UpdateNews(ctx, id, req)
	c.JSON(resp.Status, resp)
}

// DeleteNews
//
// @Summary Delete News
// @Description Delete a news article (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id} [delete]
func (h *handler) DeleteNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.DeleteNews(ctx, newsID)
	c.JSON(res.Status, res)
}

func (h *handler) UpdatePublishedNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.UpdatePublishedNews(ctx, newsID)
	c.JSON(res.Status, res)
}

func (h *handler) UpdateArchivedNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.UpdateArchivedNews(ctx, newsID)
	c.JSON(res.Status, res)
}
