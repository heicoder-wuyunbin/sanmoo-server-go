package handler

import (
	"strconv"

	domperm "sanmoo-server-go/internal/domain/permission"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	module := c.Query("module")
	ptype := c.Query("type")

	out, err := h.svc.Permission.ListPermissions(c.Request.Context(), page, size, keyword, module, ptype)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) GetPermissionTree(c *gin.Context) {
	tree, err := h.svc.Permission.PermissionTree(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.ListResponse[domperm.PermissionTree]{List: tree})
}

func (h *Handler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	perm, err := h.svc.Permission.GetPermission(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, perm)
}

func (h *Handler) CreatePermission(c *gin.Context) {
	var req dto.AdminPermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	perm := &domperm.Permission{
		PermKey:     req.PermKey,
		Name:        req.Name,
		Module:      req.Module,
		Type:        req.Type,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	id, err := h.svc.Permission.CreatePermission(c.Request.Context(), perm)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.IDResponse{ID: id})
}

func (h *Handler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminPermissionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	perm := &domperm.Permission{
		ID:          id,
		Name:        req.Name,
		Module:      req.Module,
		Type:        req.Type,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      int8(req.Status),
	}
	if err := h.svc.Permission.UpdatePermission(c.Request.Context(), perm); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Permission.DeletePermission(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}
