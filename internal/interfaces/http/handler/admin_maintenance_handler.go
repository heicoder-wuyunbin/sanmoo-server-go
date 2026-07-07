package handler

import (
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminMaintenanceCleanupLogs(c *gin.Context) {
	report, err := h.svc.Maintenance.CleanupExpiredLogs(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, report)
}

func (h *Handler) AdminMaintenanceStats(c *gin.Context) {
	stats, err := h.svc.Maintenance.GetMaintenanceStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, stats)
}