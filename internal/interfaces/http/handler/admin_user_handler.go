package handler

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	domuser "sanmoo-server-go/internal/domain/user"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/interfaces/http/middleware"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin User Management Handlers ========================

func (h *Handler) GetUsers(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.User.ListUsers(c.Request.Context(), q.Page, q.Size, q.Keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.AdminUserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	id, err := h.svc.User.CreateUser(c.Request.Context(), req.Username, req.Password, req.RoleID, req.Email)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.IDResponse{ID: id})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.UpdateUser(c.Request.Context(), id, req.Username, req.Password, req.RoleID, req.Email); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.DeleteUser(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) UpdateUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminUserUpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.UpdateUserPassword(c.Request.Context(), id, req.Password); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteUsers(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.BatchDeleteUsers(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AssignUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AssignUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Role.AssignUserRoles(c.Request.Context(), id, req.RoleIDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ToggleUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.ToggleUserStatus(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.User.ChangePassword(c.Request.Context(), userID.(uint64), req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) ExportUsers(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.User.ListUsers(c.Request.Context(), 1, 10000, q.Keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}

	list, ok := out.List.([]domuser.User)
	if !ok {
		response.Fail(c, apperr.ErrInternal)
		return
	}

	filename := fmt.Sprintf("users_%s.csv", time.Now().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"ID", "用户名", "邮箱", "昵称", "角色", "状态", "创建时间", "更新时间"}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, item := range list {
		status := "启用"
		if item.Status == "DISABLED" {
			status = "禁用"
		}
		row := []string{
			fmt.Sprintf("%d", item.ID),
			item.Username,
			item.Email,
			item.Nickname,
			item.RoleName,
			status,
			item.CreateTime.Format("2006-01-02 15:04:05"),
			item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}
}
