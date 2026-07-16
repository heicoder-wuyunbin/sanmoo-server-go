package handler

import (
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ClearCache 一键清空所有业务缓存。
func (h *Handler) ClearCache(c *gin.Context) {
	result, err := h.svc.Cache.ClearAll(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("CACHE_ERROR", err.Error()))
		return
	}
	response.Ok(c, result)
}

// WarmupCache 一键缓存所有文章、分类、标签等数据。
func (h *Handler) WarmupCache(c *gin.Context) {
	result, err := h.svc.Cache.WarmupAll(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("CACHE_ERROR", err.Error()))
		return
	}
	response.Ok(c, result)
}

// CacheStats 获取缓存统计信息。
func (h *Handler) CacheStats(c *gin.Context) {
	stats, err := h.svc.Cache.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, apperr.New("CACHE_ERROR", err.Error()))
		return
	}
	response.Ok(c, dto.MapResponse{Data: stats})
}