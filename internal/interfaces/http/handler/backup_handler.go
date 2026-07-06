package handler

import (
	"os"

	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ExportData(c *gin.Context) {
	result, err := h.svc.Backup.ExportAllData(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("BACKUP_ERROR", err.Error()))
		return
	}
	response.Ok(c, result)
}

func (h *Handler) ListBackups(c *gin.Context) {
	results, err := h.svc.Backup.ListBackups(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("BACKUP_ERROR", err.Error()))
		return
	}
	response.Ok(c, results)
}

func (h *Handler) DeleteBackup(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		response.Fail(c, apperr.New("BACKUP_ERROR", "文件名不能为空"))
		return
	}
	err := h.svc.Backup.DeleteBackup(c.Request.Context(), fileName)
	if err != nil {
		response.Fail(c, apperr.New("BACKUP_ERROR", err.Error()))
		return
	}
	response.Ok(c, gin.H{"message": "删除成功"})
}

func (h *Handler) DownloadBackup(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		response.Fail(c, apperr.New("BACKUP_ERROR", "文件名不能为空"))
		return
	}

	filePath := "./backups/" + fileName
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		response.Fail(c, apperr.New("BACKUP_ERROR", "文件不存在"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/json")
	c.File(filePath)
}

func (h *Handler) GetBackupStats(c *gin.Context) {
	backups, err := h.svc.Backup.ListBackups(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("BACKUP_ERROR", err.Error()))
		return
	}

	var totalSize int64 = 0
	for _, b := range backups {
		totalSize += b.Size
	}

	response.Ok(c, gin.H{
		"totalCount": len(backups),
		"totalSize":  totalSize,
		"backups":    backups,
	})
}
