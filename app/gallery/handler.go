package gallery

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service      Service
	mediaService media.Service
	middleware   middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, ms media.Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:      s,
		mediaService: ms,
		middleware:   m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	public := r.Group("/galleries")
	public.GET("", h.GetGalleryList)
	public.GET("/:slug", h.GetGalleryBySlug)

	admin := r.Group("/admin/galleries")
	admin.Use(h.middleware.RequireRoles(enum.RolePublicationManager))
	{
		admin.GET("", h.GetAdminGalleryList)
		admin.GET("/:id", h.GetGalleryByID)
		admin.POST("", h.CreateGallery)
		admin.PUT("/:id", h.UpdateGallery)
		admin.DELETE("/:id", h.DeleteGallery)
		admin.PATCH("/:id/publish", h.UpdatePublishGallery)
		admin.PATCH("/:id/archive", h.UpdateArchiveGallery)
	}
}

// GetGalleryList
//
// @Summary List Published Galleries
// @Description Retrieve a list of published gallery items with offset-based pagination and optional filters
// @Tags Gallery
// @Accept json
// @Produce json
// @Param search query string false "Search gallery by title"
// @Param category query string false "Filter by category"
// @Param sortBy query string false "Sort by field (e.g. 'created_at desc', 'views desc', 'title asc')"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=GalleryListResponse}
// @Router /api/public/galleries/ [get]
func (h *handler) GetGalleryList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams GalleryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetGalleryList(ctx, queryParams, false)
	c.JSON(res.Status, res)
}

// GetGalleryBySlug
//
// @Summary Get Published Gallery
// @Description Get detailed information of a specific published gallery item
// @Tags Gallery
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response{data=GalleryResponse}
// @Router /api/public/galleries/{id} [get]
func (h *handler) GetGalleryBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	gallerySlug := c.Param("slug")

	res := h.service.GetGalleryBySlug(ctx, gallerySlug)
	c.JSON(res.Status, res)
}

// GetAdminGalleryList
//
// @Summary List All Galleries (Protected)
// @Description Retrieve a list of all gallery items (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search gallery by title"
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Param sortBy query string false "Sort by field (e.g. 'created_at desc', 'views desc', 'title asc')"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=GalleryListResponse}
// @Router /api/galleries/ [get]
func (h *handler) GetAdminGalleryList(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams GalleryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetGalleryList(ctx, queryParams, true)
	c.JSON(res.Status, res)
}

// GetGalleryByID
//
// @Summary Get Gallery (Protected)
// @Description Get detailed information of a specific gallery item (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response{data=GalleryResponse}
// @Router /api/galleries/{id} [get]
func (h *handler) GetGalleryByID(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.GetGalleryByID(ctx, galleryID)
	c.JSON(res.Status, res)
}

// CreateGallery
//
// @Summary Create Gallery
// @Description Create a new gallery item (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Gallery Title"
// @Param category formData string true "Gallery Category"
// @Param description formData string true "Gallery Description"
// @Param status formData string true "Gallery Status (draft, published, archived)"
// @Param coverImage formData file false "Gallery Cover Image"
// @Param mediaFiles[] formData file false "Gallery Media Files"
// @Param mediaAlt[] formData string false "Media Alt Texts"
// @Success 201 {object} pkg.Response{data=GalleryResponse}
// @Router /api/galleries/ [post]
func (h *handler) CreateGallery(c *gin.Context) {
	ctx := c.Request.Context()

	var req GalleryCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.CreateGallery(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateGallery
//
// @Summary Update Gallery
// @Description Update an existing gallery item (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Gallery ID"
// @Param title formData string false "Gallery Title"
// @Param category formData string false "Gallery Category"
// @Param description formData string false "Gallery Description"
// @Param status formData string false "Gallery Status"
// @Param coverImage formData file false "Gallery Cover Image"
// @Param mediaFiles[] formData file false "Gallery Media Files"
// @Param mediaAlt[] formData string false "Media Alt Texts"
// @Success 200 {object} pkg.Response
// @Router /api/galleries/{id} [put]
func (h *handler) UpdateGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	var req GalleryUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.UpdateGallery(ctx, galleryID, req)
	c.JSON(res.Status, res)
}

// DeleteGallery
//
// @Summary Delete Gallery
// @Description Delete a gallery item (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response
// @Router /api/galleries/{id} [delete]
func (h *handler) DeleteGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.DeleteGallery(ctx, galleryID)
	c.JSON(res.Status, res)
}

// UpdatePublishGallery
//
// @Summary Update Publish Gallery
// @Description Update an existing gallery to publish (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/galleries/{id}/publish [patch]
func (h *handler) UpdatePublishGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.UpdatePublishGallery(ctx, galleryID)
	c.JSON(res.Status, res)
}

// UpdateArchiveGallery
//
// @Summary Update Archive Gallery
// @Description Update an existing gallery to archived (requires publication manager or superadmin role)
// @Tags Gallery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/galleries/{id}/archive [patch]
func (h *handler) UpdateArchiveGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.UpdateArchivedGallery(ctx, galleryID)
	c.JSON(res.Status, res)
}
