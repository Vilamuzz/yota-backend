package news

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/minio"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service     Service
	minioClient minio.Client
	middleware  middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, minioClient minio.Client, m middleware.AppMiddleware) {
	handler := &handler{
		service:     s,
		minioClient: minioClient,
		middleware:  m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	api := r.Group("/news")

	// Public routes
	api.GET("/", h.GetAllNews)
	api.GET("/:id", h.GetNewsByID)

	// Protected routes (require publication manager role)
	protected := api.Group("")
	protected.Use(h.middleware.RequireRoles(string(user.RolePublicationManager), string(user.RoleSuperadmin)))
	{
		protected.POST("/", h.CreateNews)
		protected.PUT("/:id", h.UpdateNews)
		protected.DELETE("/:id", h.DeleteNews)
	}
}

// GetAllNews
//
// @Summary Get All News
// @Description Retrieve a list of all news with cursor-based pagination and optional filters
// @Tags News
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (general, event, announcement, donation, social)"
// @Param status query string false "Filter by status (draft, published, archived)"
// @Param cursor query string false "Cursor for pagination (encoded string)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} pkg.Response
// @Router /api/news/ [get]
func (h *handler) GetAllNews(c *gin.Context) {
	ctx := c.Request.Context()

	var queryParams NewsQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.FetchAllNews(ctx, queryParams)
	c.JSON(res.Status, res)
}

// GetNewsByID
//
// @Summary Get News by ID
// @Description Get detailed information of a specific news article
// @Tags News
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} pkg.Response
// @Router /api/news/{id} [get]
func (h *handler) GetNewsByID(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	// Increment view count for public access
	res := h.service.FetchNewsByID(ctx, newsID, true)
	c.JSON(res.Status, res)
}

// CreateNews
//
// @Summary Create News
// @Description Create a new news article (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "News Title"
// @Param category formData string true "News Category"
// @Param content formData string true "News Content"
// @Param image formData file false "News Image File"
// @Param image_url formData string false "News Image URL (if not uploading file)"
// @Param status formData string false "News Status"
// @Success 201 {object} pkg.Response
// @Router /api/news/ [post]
func (h *handler) CreateNews(c *gin.Context) {
	ctx := c.Request.Context()

	var req NewsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil {
		// File uploaded, save to MinIO
		fileURL, err := h.minioClient.UploadFile(ctx, file, "news")
		if err != nil {
			c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil))
			return
		}
		req.Image = fileURL
	}

	res := h.service.CreateNews(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateNews
//
// @Summary Update News
// @Description Update an existing news article (requires publication manager or superadmin role)
// @Tags News
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "News ID"
// @Param title formData string false "News Title"
// @Param category formData string false "News Category"
// @Param content formData string false "News Content"
// @Param image formData file false "News Image File"
// @Param image_url formData string false "News Image URL (if not uploading file)"
// @Param status formData string false "News Status"
// @Success 200 {object} pkg.Response
// @Router /api/news/{id} [put]
func (h *handler) UpdateNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	var req UpdateNewsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil {
		// File uploaded, save to MinIO
		fileURL, err := h.minioClient.UploadFile(ctx, file, "news")
		if err != nil {
			c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil))
			return
		}
		req.Image = fileURL
	}

	res := h.service.UpdateNews(ctx, newsID, req)
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
// @Router /api/news/{id} [delete]
func (h *handler) DeleteNews(c *gin.Context) {
	ctx := c.Request.Context()
	newsID := c.Param("id")

	res := h.service.DeleteNews(ctx, newsID)
	c.JSON(res.Status, res)
}
