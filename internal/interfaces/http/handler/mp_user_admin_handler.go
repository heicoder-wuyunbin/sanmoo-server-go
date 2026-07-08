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
//
// FROZEN (L2): 该接口属于"重运营画像"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再主动生成六边形画像。
// 接口签名保留以兼容现有管理端调用，避免前端直接崩溃；
// 实际计算逻辑已在 service 层短路，返回空画像结构。
// 详见 documents/mp-user-domain-downgrade.md。
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
//
// FROZEN (L2): 该接口属于"重运营标签"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再自动生成行为/兴趣标签。
// 接口签名保留以兼容现有管理端调用，避免前端直接崩溃；
// 实际计算逻辑已在 service 层短路，返回空标签列表。
// 详见 documents/mp-user-domain-downgrade.md。
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
//
// FROZEN (L2): 该接口属于"雷达画像生成"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再聚合生成雷达画像。
// 接口签名保留以兼容现有管理端调用，避免前端直接崩溃；
// 实际计算逻辑已在 service 层短路，返回空雷达结构。
// 详见 documents/mp-user-domain-downgrade.md。
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
