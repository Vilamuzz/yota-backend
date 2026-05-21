package news_comment

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

func NewHandler(r *gin.RouterGroup, service Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:    service,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/news/:slug/comments", h.middleware.AuthOptional(), h.GetNewsCommentList)
	router.GET("/news/comments/:id", h.middleware.AuthOptional(), h.GetNewsCommentByID)
	router.POST("/news/:slug/comments", h.middleware.AuthRequired(), h.CreateNewsComment)

	publicProtected := router.Group("/news/comments")
	publicProtected.Use(h.middleware.AuthRequired())
	{
		publicProtected.POST("/:id/report", h.CreateReportNewsComment)
	}

	admin := router.Group("/admin/news/comments")
	admin.Use(h.middleware.RequireRoles(enum.RolePublicationManager))
	{
		admin.GET("", h.GetReportedNewsCommentList)
		admin.PATCH("/:id/allow", h.AllowNewsComment)
		admin.DELETE("/:id", h.DeleteNewsComment)
	}
}

// @Summary Report Prayer
// @Description Report a prayer
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Param body body ReportPrayerRequest true "Report Prayer Payload"
// @Success 200 {object} pkg.Response
// @Router /api/prayers/{id}/report [post]
func (h *handler) CreateReportNewsComment(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	accountID := claims.AccountID
	var payload ReportNewsCommentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	newsCommentID := c.Param("id")
	res := h.service.CreateReportNewsComment(ctx, newsCommentID, accountID, payload)
	c.JSON(res.Status, res)
}

func (h *handler) AllowNewsComment(c *gin.Context) {
	ctx := c.Request.Context()
	newsCommentID := c.Param("id")
	res := h.service.AllowNewsComment(ctx, newsCommentID)
	c.JSON(res.Status, res)
}

func (h *handler) CreateNewsComment(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	accountID := claims.AccountID
	var payload CreateNewsCommentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	slug := c.Param("slug")
	res := h.service.CreateNewsComment(ctx, slug, accountID, payload)
	c.JSON(res.Status, res)
}

// @Summary Get Prayer by ID
// @Description Get a prayer by its ID
// @Tags Prayer
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response{data=PrayerResponse}
// @Router /api/prayers/{id} [get]
func (h *handler) GetNewsCommentByID(c *gin.Context) {
	ctx := c.Request.Context()
	newsCommentID := c.Param("id")
	accountID := ""
	if accountData, exists := c.Get("user_data"); exists {
		if claims, ok := accountData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}
	res := h.service.GetNewsCommentByID(ctx, newsCommentID, accountID)
	c.JSON(res.Status, res)
}

// @Summary List Prayers
// @Description Get a list of prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Router /api/donation-programs/{slug}/prayers [get]
func (h *handler) GetNewsCommentList(c *gin.Context) {
	ctx := c.Request.Context()
	newsSlug := c.Param("slug")
	var params NewsCommentQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	accountID := ""
	if accountData, exists := c.Get("user_data"); exists {
		if claims, ok := accountData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}

	res := h.service.GetNewsCommentList(ctx, accountID, newsSlug, false, params)
	c.JSON(res.Status, res)
}

// @Summary List Reported Prayers
// @Description Get a list of reported prayers
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Router /api/admin/prayers [get]
func (h *handler) GetReportedNewsCommentList(c *gin.Context) {
	ctx := c.Request.Context()
	var params NewsCommentQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.GetNewsCommentList(ctx, "", "", true, params)
	c.JSON(res.Status, res)
}

// @Summary Delete Prayer
// @Description Delete a prayer by its ID
// @Tags Prayer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/prayers/{id} [delete]
func (h *handler) DeleteNewsComment(c *gin.Context) {
	ctx := c.Request.Context()
	newsCommentID := c.Param("id")
	res := h.service.DeleteNewsComment(ctx, newsCommentID)
	c.JSON(res.Status, res)
}
