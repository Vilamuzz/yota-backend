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
	public.GET("", h.GetNewsList)
	public.GET("/:slug", h.GetNewsBySlug)

	admin := r.Group("/admin/news")
	admin.Use(h.middleware.RequireRoles(enum.RolePublicationManager))
	{
		admin.GET("", h.GetAdminNewsList)
		admin.GET("/:id", h.GetNewsByID)
		admin.POST("", h.CreateNews)
		admin.PUT("/:id", h.UpdateNews)
		admin.DELETE("/:id", h.DeleteNews)
		admin.PATCH("/:id/publish", h.UpdatePublishNews)
		admin.PATCH("/:id/archive", h.UpdateArchiveNews)
	}
}

// GetNewsList
//
// @Summary List Published News
// @Description Retrieve a list of published news items with offset-based pagination and optional filters
// @Tags News
// @Accept json
// @Produce json
// @Param search query string false "Search news by title"
// @Param category query string false "Filter by category"
// @Param sortBy query string false "Sort by field (e.g. 'created_at desc', 'views desc', 'published_at desc', 'title asc')"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=NewsListResponse}
// @Router /api/news/ [get]
func (h *handler) GetNewsList(c *gin.Context) {
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
// @Success 200 {object} pkg.Response{data=NewsResponse}
// @Router /api/news/{slug} [get]
func (h *handler) GetNewsBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	res := h.service.GetNewsBySlug(ctx, slug)
	c.JSON(res.Status, res)
}

// GetAdminNewsList
//
// @Summary List All News (Protected)
// @Description Retrieve a list of all news items (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search news by title"
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Param sortBy query string false "Sort by field (e.g. 'created_at desc', 'views desc', 'published_at desc', 'title asc')"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=NewsListResponse}
// @Router /api/admin/news/ [get]
func (h *handler) GetAdminNewsList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams NewsQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetNewsList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

// GetNewsByID
//
// @Summary Get News (Protected)
// @Description Get detailed information of a specific news article (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response{data=NewsResponse}
// @Router /api/admin/news/{id} [get]
func (h *handler) GetNewsByID(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.GetNewsByID(ctx, newsID)
	c.JSON(res.Status, res)
}

// CreateNews
//
// @Summary Create News
// @Description Create a new news item (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "News Title"
// @Param category formData string true "News Category"
// @Param content formData string true "News Content"
// @Param status formData string true "News Status (draft, published, archived)"
// @Param coverImage formData file false "News Cover Image"
// @Param mediaFiles[] formData file false "News Media Files"
// @Param mediaAlt[] formData string false "Media Alt Texts"
// @Success 201 {object} pkg.Response{data=NewsResponse}
// @Router /api/admin/news [post]
func (h *handler) CreateNews(c *gin.Context) {
	ctx := c.Request.Context()

	var req NewsCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}
	res := h.service.CreateNews(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateNews
//
// @Summary Update News
// @Description Update an existing news item (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "News ID"
// @Param title formData string false "News Title"
// @Param category formData string false "News Category"
// @Param content formData string false "News Content"
// @Param status formData string false "News Status"
// @Param coverImage formData file false "News Cover Image"
// @Param mediaFiles[] formData file false "News Media Files"
// @Param mediaAlt[] formData string false "Media Alt Texts"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id} [put]
func (h *handler) UpdateNews(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req NewsUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}
	res := h.service.UpdateNews(ctx, id, req)
	c.JSON(res.Status, res)
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

// UpdatePublishNews
//
// @Summary Update Publish News
// @Description Update an existing news to publish (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id}/publish [patch]
func (h *handler) UpdatePublishNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.UpdatePublishNews(ctx, newsID)
	c.JSON(res.Status, res)
}

// UpdateArchiveNews
//
// @Summary Update Archive News
// @Description Update an existing news to archived (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/news/{id}/archive [patch]
func (h *handler) UpdateArchiveNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.UpdateArchiveNews(ctx, newsID)
	c.JSON(res.Status, res)
}
