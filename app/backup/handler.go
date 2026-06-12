package backup

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
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
	backupGroup := r.Group("admin/backups")
	backupGroup.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		backupGroup.POST("", h.CreateBackup)
		backupGroup.GET("", h.ListBackups)
		backupGroup.GET("/:id/download", h.GetBackupURL)
		backupGroup.DELETE("/:id", h.DeleteBackup)
		backupGroup.POST("/cleanup", h.CleanupOldBackups)
	}
}

// CreateBackup handles on-demand backup creation
// @Summary Create database backup
// @Description Create an immediate backup of the database and store in S3 and register in metadata db
// @Tags Backup
// @Security BearerAuth
// @Produce json
// @Success 201 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/admin/backups [post]
func (h *handler) CreateBackup(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.CreateBackup(ctx)
	c.JSON(res.Status, res)
}

// ListBackups handles listing all backups
// @Summary List all backups
// @Description List all database backups stored in database metadata
// @Tags Backup
// @Security BearerAuth
// @Produce json
// @Success 200 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/admin/backups [get]
func (h *handler) ListBackups(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.ListBackups(ctx)
	c.JSON(res.Status, res)
}

// GetBackupURL handles getting a presigned download URL
// @Summary Get backup download URL
// @Description Get presigned URL for downloading a backup
// @Tags Backup
// @Security BearerAuth
// @Param id path string true "Backup ID"
// @Produce json
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/admin/backups/{id}/download [get]
func (h *handler) GetBackupURL(c *gin.Context) {
	ctx := c.Request.Context()
	backupID := c.Param("id")
	res := h.service.GetBackupURL(ctx, backupID)
	c.JSON(res.Status, res)
}

// DeleteBackup handles deleting a backup
// @Summary Delete backup
// @Description Delete a backup from S3 and metadata database
// @Tags Backup
// @Security BearerAuth
// @Param id path string true "Backup ID"
// @Produce json
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/admin/backups/{id} [delete]
func (h *handler) DeleteBackup(c *gin.Context) {
	ctx := c.Request.Context()
	backupID := c.Param("id")
	res := h.service.DeleteBackup(ctx, backupID)
	c.JSON(res.Status, res)
}

// CleanupOldBackups handles cleanup of backups older than retention period
// @Summary Cleanup old backups
// @Description Manually trigger cleanup of backups older than specified days
// @Tags Backup
// @Security BearerAuth
// @Param retention query integer false "Retention days (default: 7)"
// @Produce json
// @Success 200 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/admin/backups/cleanup [post]
func (h *handler) CleanupOldBackups(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParams CleanupQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	retentionDays := 7
	if queryParams.RetentionDays != nil {
		retentionDays = *queryParams.RetentionDays
	}

	res := h.service.CleanupOldBackups(ctx, retentionDays)
	c.JSON(res.Status, res)
}
