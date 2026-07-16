package handler

import (
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin MP User Management Handlers ========================

// 轻运营：仅保留用户列表查询，重运营画像/标签/雷达图功能已移除。
// 详见 documents/mp-user-domain-downgrade.md

func (h *Handler) GetMPUsers(c *gin.Context) {
	var q dto.MPUserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.MPUser.ListMPUsers(c.Request.Context(), q.Page, q.Size, q.Keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}