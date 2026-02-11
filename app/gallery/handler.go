package gallery

import (
	"encoding/json"
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
	api := r.Group("/galleries")

	// Public routes
	api.GET("/", h.GetAllGalleries)
	api.GET("/:id", h.GetGalleryByID)

	// Protected routes (require publication manager or superadmin role)
	protected := api.Group("")
	protected.Use(h.middleware.RequireRoles(string(user.RolePublicationManager), string(user.RoleSuperadmin)))
	{
		protected.POST("/", h.CreateGallery)
		protected.PUT("/:id", h.UpdateGallery)
		protected.DELETE("/:id", h.DeleteGallery)
	}
}

// GetAllGalleries
//
// @Summary Get All Galleries
// @Description Retrieve a list of all gallery items with cursor-based pagination and optional filters
// @Tags Gallery
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (photography, painting, sculpture, digital, mixed)"
// @Param status query string false "Filter by status (active, inactive, archived)"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/galleries/ [get]
func (h *handler) GetAllGalleries(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams GalleryQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.FetchAllGalleries(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetGalleryByID
//
// @Summary Get Gallery by ID
// @Description Get detailed information of a specific gallery item
// @Tags Gallery
// @Accept json
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} pkg.Response
// @Router /api/galleries/{id} [get]
func (h *handler) GetGalleryByID(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	// Increment view count for public access
	res := h.service.FetchGalleryByID(ctx, galleryID, true)
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
// @Param status formData string false "Gallery Status"
// @Param files formData file true "Media Files (can be multiple)"
// @Success 201 {object} pkg.Response
// @Router /api/galleries/ [post]
func (h *handler) CreateGallery(c *gin.Context) {
	ctx := c.Request.Context()

	var req GalleryRequest
	// Attempt to bind multipart/form-data or JSON
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Handle file upload
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["files"]
		mediaItems, err := h.mediaService.UploadMedia(ctx, files, "galleries")
		if err != nil {
			c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload file", nil, nil))
			return
		}
		req.Media = append(req.Media, mediaItems...)
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
// @Param existing_media formData string false "Existing media JSON array"
// @Param files formData file false "Media Files (can be multiple)"
// @Success 200 {object} pkg.Response
// @Router /api/galleries/{id} [put]
func (h *handler) UpdateGallery(c *gin.Context) {
	ctx := c.Request.Context()
	galleryID := c.Param("id")

	var req UpdateGalleryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Parse existing_media if provided
	existingMediaJSON := c.PostForm("existing_media")
	if existingMediaJSON != "" {
		var existingMedia []media.MediaRequest
		if err := json.Unmarshal([]byte(existingMediaJSON), &existingMedia); err != nil {
			c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid existing_media format", nil, nil))
			return
		}
		req.Media = existingMedia
	}

	// Handle new file uploads
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["files"]
		if len(files) > 0 {
			// Upload new files
			newMediaItems, err := h.mediaService.UploadMedia(ctx, files, "galleries")
			if err != nil {
				c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload files", nil, nil))
				return
			}
			// Append new media to existing media
			req.Media = append(req.Media, newMediaItems...)
		}
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
