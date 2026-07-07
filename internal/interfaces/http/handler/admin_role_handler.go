package handler

import (
	"strconv"

	domrole "sanmoo-server-go/internal/domain/role"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/interfaces/http/middleware"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	out, err := h.svc.Role.ListRoles(c.Request.Context(), page, size, keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) GetAllRoles(c *gin.Context) {
	out, err := h.svc.Role.ListAllRoles(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	role, err := h.svc.Role.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, role)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req dto.AdminRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	role := &domrole.Role{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	id, err := h.svc.Role.CreateRole(c.Request.Context(), role)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.IDResponse{ID: id})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminRoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	role := &domrole.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      int8(req.Status),
	}
	if err := h.svc.Role.UpdateRole(c.Request.Context(), role); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Role.DeleteRole(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AssignRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Role.AssignPermissions(c.Request.Context(), id, req.PermKeys); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	keys, err := h.svc.Role.GetRolePermKeys(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, gin.H{"permKeys": keys})
}

func (h *Handler) GetUserPermissions(c *gin.Context) {
	userID, ok := c.Get(middleware.CtxUserIDKey)
	if !ok {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	uid := userID.(uint64)
	permSet, err := h.svc.Role.GetUserPermKeys(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	keys := make([]string, 0, len(permSet))
	for k := range permSet {
		keys = append(keys, k)
	}
	response.Ok(c, gin.H{"permKeys": keys})
}

func (h *Handler) GetUserMenus(c *gin.Context) {
	userID, ok := c.Get(middleware.CtxUserIDKey)
	if !ok {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	uid := userID.(uint64)
	menus, err := h.svc.Role.GetUserMenus(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, gin.H{"menus": menus})
}
