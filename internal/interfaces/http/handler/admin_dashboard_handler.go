package handler

import (
	"strconv"

	domdashboard "sanmoo-server-go/internal/domain/dashboard"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin Dashboard Handlers ========================

func (h *Handler) Dashboard(c *gin.Context) {
	out, err := h.svc.Dashboard.Dashboard(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) VisitorRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	out, err := h.svc.Dashboard.VisitorRecords(c.Request.Context(), page, size, keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) ErrorLogRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	out, err := h.svc.Dashboard.ErrorLogRecords(c.Request.Context(), page, size, keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) DeleteVisitorRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Dashboard.DeleteVisitorRecord(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteVisitorRecords(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Dashboard.BatchDeleteVisitorRecords(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ClearAllVisitorRecords(c *gin.Context) {
	if err := h.svc.Dashboard.ClearAllVisitorRecords(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeleteErrorLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Dashboard.DeleteErrorLog(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteErrorLogs(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Dashboard.BatchDeleteErrorLogs(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ClearAllErrorLogs(c *gin.Context) {
	if err := h.svc.Dashboard.ClearAllErrorLogs(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ImportErrorLogs(c *gin.Context) {
	var req dto.ImportErrorLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if len(req.Logs) == 0 {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	logs := make([]domdashboard.ErrorLogRecord, 0, len(req.Logs))
	for _, item := range req.Logs {
		logs = append(logs, domdashboard.ErrorLogRecord{
			ErrorCode:     item.ErrorCode,
			ErrorMessage:  item.ErrorMessage,
			ErrorDetail:   item.ErrorDetail,
			StackTrace:    item.StackTrace,
			RequestURL:    item.RequestURL,
			RequestMethod: item.RequestMethod,
			RequestParams: item.RequestParams,
			RequestBody:   item.RequestBody,
			ResponseBody:  item.ResponseBody,
			IPAddress:     item.IPAddress,
			UserAgent:     item.UserAgent,
		})
	}
	count, err := h.svc.Dashboard.ImportErrorLogs(c.Request.Context(), logs)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, map[string]int64{"imported": count})
}

func (h *Handler) ExportErrorLogs(c *gin.Context) {
	logs, err := h.svc.Dashboard.ExportErrorLogs(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, logs)
}

func (h *Handler) PV(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	out, err := h.svc.Dashboard.PV(c.Request.Context(), days)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) TagStatistics(c *gin.Context) {
	out, err := h.svc.Dashboard.TagStatistics(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) CategoryStatistics(c *gin.Context) {
	out, err := h.svc.Dashboard.CategoryStatistics(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) ArticlePublishHeatmap(c *gin.Context) {
	out, err := h.svc.Dashboard.ArticlePublishHeatmap(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) TopicStatistics(c *gin.Context) {
	out, err := h.svc.Dashboard.TopicStatistics(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) MpUserGrowth(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	out, err := h.svc.Dashboard.MpUserGrowth(c.Request.Context(), days)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) ArticleReadStatistics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	out, err := h.svc.Dashboard.ArticleReadStatistics(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) CategoryReadStatistics(c *gin.Context) {
	out, err := h.svc.Dashboard.CategoryReadStatistics(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) TagReadStatistics(c *gin.Context) {
	out, err := h.svc.Dashboard.TagReadStatistics(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) ContentTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	out, err := h.svc.Dashboard.ContentTrend(c.Request.Context(), days)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}
