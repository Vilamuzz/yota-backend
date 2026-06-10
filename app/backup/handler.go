package backup

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/Vilamuzz/yota-backend/app/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(router *gin.RouterGroup, service Service, mw middleware.AppMiddleware) {
	h := &Handler{service: service}

	backupGroup := router.Group("/backups")
	backupGroup.Use(mw.AuthRequired())
	{
		backupGroup.POST("/create", h.CreateBackup)
		backupGroup.GET("/list", h.ListBackups)
		backupGroup.GET("/download/:id", h.GetBackupURL)
		backupGroup.DELETE("/:id", h.DeleteBackup)
		backupGroup.POST("/cleanup", h.CleanupOldBackups)
	}
}

// CreateBackup handles on-demand backup creation
// @Summary Create database backup
// @Description Create an immediate backup of the database and store in S3
// @Tags Backup
// @Param backup body BackupRequest false "Backup request"
// @Success 200 {object} BackupResponse
// @Failure 400 {object} BackupResponse
// @Failure 500 {object} BackupResponse
// @Router /backups/create [post]
func (h *Handler) CreateBackup(c *gin.Context) {
	ctx := c.Request.Context()

	backup, err := h.service.CreateBackup(ctx)
	if err != nil {
		logrus.Errorf("Handler: failed to create backup: %v", err)
		c.JSON(http.StatusInternalServerError, BackupResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BackupResponse{
		Success: true,
		Message: "Backup created successfully",
		Backup:  backup,
	})
}

// ListBackups handles listing all backups
// @Summary List all backups
// @Description List all database backups stored in S3
// @Tags Backup
// @Success 200 {object} BackupListResponse
// @Failure 500 {object} BackupListResponse
// @Router /backups/list [get]
func (h *Handler) ListBackups(c *gin.Context) {
	ctx := c.Request.Context()

	backups, err := h.service.ListBackups(ctx)
	if err != nil {
		logrus.Errorf("Handler: failed to list backups: %v", err)
		c.JSON(http.StatusInternalServerError, BackupListResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BackupListResponse{
		Success: true,
		Backups: backups,
		Total:   len(backups),
	})
}

// GetBackupURL handles getting a presigned download URL
// @Summary Get backup download URL
// @Description Get presigned URL for downloading a backup
// @Tags Backup
// @Param id path string true "Backup ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /backups/download/{id} [get]
func (h *Handler) GetBackupURL(c *gin.Context) {
	backupID := c.Param("id")
	ctx := c.Request.Context()

	url, err := h.service.GetBackupURL(ctx, backupID)
	if err != nil {
		logrus.Errorf("Handler: failed to get backup URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     url,
	})
}

// DeleteBackup handles deleting a backup
// @Summary Delete backup
// @Description Delete a backup from S3
// @Tags Backup
// @Param id path string true "Backup ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /backups/{id} [delete]
func (h *Handler) DeleteBackup(c *gin.Context) {
	backupID := c.Param("id")
	ctx := c.Request.Context()

	err := h.service.DeleteBackup(ctx, backupID)
	if err != nil {
		logrus.Errorf("Handler: failed to delete backup: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Backup deleted successfully",
	})
}

// CleanupOldBackups handles cleanup of backups older than retention period
// @Summary Cleanup old backups
// @Description Manually trigger cleanup of backups older than specified days
// @Tags Backup
// @Param retention query integer false "Retention days (default: 7)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /backups/cleanup [post]
func (h *Handler) CleanupOldBackups(c *gin.Context) {
	retentionDays := 7
	if days := c.Query("retention"); days != "" {
		if _, err := fmt.Sscanf(days, "%d", &retentionDays); err != nil {
			logrus.Warnf("Handler: invalid retention days value, using default: %v", err)
			retentionDays = 7
		}
	}

	ctx := c.Request.Context()
	deletedCount, err := h.service.CleanupOldBackups(ctx, retentionDays)
	if err != nil {
		logrus.Errorf("Handler: failed to cleanup old backups: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Cleanup completed successfully",
		"deleted_count": deletedCount,
		"retention_days": retentionDays,
	})
}
