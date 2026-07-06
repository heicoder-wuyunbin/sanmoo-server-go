package handler

import (
	"strconv"

	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin MP User Management Handlers ========================

// AdminMPUsers 微信用户列表。
func (h *Handler) AdminMPUsers(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")
	tagName := c.Query("tagName")

	out, err := h.svc.MPUser.ListMPUsers(c.Request.Context(), page, size, keyword, tagName)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserDetail 微信用户详情。
func (h *Handler) AdminMPUserDetail(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.GetMPUserDetail(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserProfile 微信用户六边形画像。
func (h *Handler) AdminMPUserProfile(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.GetMPUserProfile(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserGenerateProfile 生成用户六边形画像。
func (h *Handler) AdminMPUserGenerateProfile(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.GenerateUserProfile(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserGenerateTags 自动生成用户标签。
func (h *Handler) AdminMPUserGenerateTags(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.GenerateUserTags(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserTags 用户标签列表。
func (h *Handler) AdminMPUserTags(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.GetUserTags(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// AdminMPUserDeleteTag 删除用户标签。
func (h *Handler) AdminMPUserDeleteTag(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.Param("tagId"), 10, 64)
	if err != nil || tagID == 0 {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.MPUser.DeleteUserTag(c.Request.Context(), tagID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

// AdminMPUserRefreshRadar 刷新雷达图（行为标签 + 兴趣维度 + 六边形画像）。
func (h *Handler) AdminMPUserRefreshRadar(c *gin.Context) {
	openID := c.Param("openid")
	if openID == "" {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.RefreshRadar(c.Request.Context(), openID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}
