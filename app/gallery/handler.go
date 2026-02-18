package gallery

import (
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
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
	// Public routes
	public := r.Group("/public/galleries")
	public.GET("", h.ListPublishedGalleries)
	public.GET("/:id", h.GetPublishedGallery)

	// Protected routes (require publication manager or superadmin role)
	protected := r.Group("/galleries")
	protected.Use(h.middleware.RequireRoles(string(user.RolePublicationManager), string(user.RoleSuperadmin)))
	{
		protected.GET("", h.ListGalleries)
		protected.GET("/:id", h.GetGallery)
		protected.POST("", h.CreateGallery)
		protected.PUT("/:id", h.UpdateGallery)
		protected.DELETE("/:id", h.DeleteGallery)
	}
}

// ListPublishedGalleries
//
// @Summary List Published Galleries
// @Description Retrieve a list of published gallery items with cursor-based pagination and optional filters
// @Tags Gallery (Public)
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=PublishedGalleryListResponse}
// @Router /public/galleries/ [get]
func (h *handler) ListPublishedGalleries(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams GalleryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.ListPublished(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetPublishedGallery
//
// @Summary Get Published Gallery
// @Description Get detailed information of a specific published gallery item
// @Tags Gallery (Public)
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response{data=PublishedGalleryResponse}
// @Router /public/galleries/{id} [get]
func (h *handler) GetPublishedGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	// Increment view count for public access
	res := h.service.GetPublishedByID(ctx, galleryID, true)
	c.JSON(res.Status, res)
}

// ListGalleries
//
// @Summary List All Galleries (Admin)
// @Description Retrieve a list of all gallery items (requires publication manager or superadmin role)
// @Tags Gallery (Admin)
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response{data=PublishedGalleryListResponse}
// @Router /galleries/ [get]
func (h *handler) ListGalleries(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams GalleryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.List(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetGallery
//
// @Summary Get Gallery (Admin)
// @Description Get detailed information of a specific gallery item (requires publication manager or superadmin role)
// @Tags Gallery (Admin)
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response{data=GalleryResponse}
// @Router /galleries/{id} [get]
func (h *handler) GetGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.GetByID(ctx, galleryID)
	c.JSON(res.Status, res)
}

// CreateGallery
//
// @Summary Create Gallery
// @Description Create a new gallery item (requires publication manager or superadmin role)
// @Tags Gallery (Admin)
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Gallery Title"
// @Param category_id formData int true "Gallery Category ID"
// @Param description formData string true "Gallery Description"
// @Param published formData boolean true "Published Status"
// @Param metadata formData string false "Media metadata JSON (array of objects with alt_text and order)"
// @Param files formData file true "Media Files (can be multiple)"
// @Success 201 {object} pkg.Response{data=GalleryResponse}
// @Router /galleries/ [post]
func (h *handler) CreateGallery(c *gin.Context) {
	ctx := c.Request.Context()

	var req GalleryRequest
	// Attempt to bind multipart/form-data or JSON
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Parse metadata if provided
	var metadataWrapper *media.MetadataWrapper
	if req.Metadata != "" {
		metadataWrapper = &media.MetadataWrapper{}
		if err := json.Unmarshal([]byte(req.Metadata), metadataWrapper); err != nil {
			c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid metadata format", nil, nil))
			return
		}
	}

	// Get uploaded files
	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["files"]
	}

	res := h.service.CreateGallery(ctx, req, files, metadataWrapper)
	c.JSON(res.Status, res)
}

// UpdateGallery
//
// @Summary Update Gallery
// @Description Update an existing gallery item (requires publication manager or superadmin role)
// @Tags Gallery (Admin)
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Gallery ID"
// @Param title formData string false "Gallery Title"
// @Param category_id formData int false "Gallery Category ID"
// @Param description formData string false "Gallery Description"
// @Param published formData boolean false "Published Status"
// @Param metadata formData string false "Media metadata JSON (array of objects with id, alt_text, and order)"
// @Param existing_media formData string false "Existing media JSON array (deprecated, use metadata instead)"
// @Param files formData file false "Media Files (can be multiple)"
// @Success 200 {object} pkg.Response
// @Router /galleries/{id} [put]
func (h *handler) UpdateGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	var req UpdateGalleryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Parse metadata if provided
	var metadataWrapper *media.MetadataWrapper
	if req.Metadata != "" {
		metadataWrapper = &media.MetadataWrapper{}
		if err := json.Unmarshal([]byte(req.Metadata), metadataWrapper); err != nil {
			c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid metadata format", nil, nil))
			return
		}
	}

	// Parse existing_media if provided (for backward compatibility)
	existingMediaJSON := c.PostForm("existing_media")
	if existingMediaJSON != "" {
		var existingMedia []media.MediaRequest
		if err := json.Unmarshal([]byte(existingMediaJSON), &existingMedia); err != nil {
			c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid existing_media format", nil, nil))
			return
		}
		req.Media = existingMedia
	}

	// Get uploaded files
	form, _ := c.MultipartForm()
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["files"]
	}

	res := h.service.UpdateGallery(ctx, galleryID, req, files, metadataWrapper)
	c.JSON(res.Status, res)
}

// DeleteGallery
//
// @Summary Delete Gallery
// @Description Delete a gallery item (requires publication manager or superadmin role)
// @Tags Gallery (Admin)
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response
// @Router /galleries/{id} [delete]
func (h *handler) DeleteGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	res := h.service.DeleteGallery(ctx, galleryID)
	c.JSON(res.Status, res)
}
