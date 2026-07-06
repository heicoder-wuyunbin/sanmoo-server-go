package handler

import (
	"strconv"

	"sanmoo-server-go/internal/domain/link"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLinks(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 {
		size = 10
	}
	keyword := c.Query("keyword")

	list, total, err := h.svc.Link.List(page, size, keyword)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.PageResponse[*link.Link]{List: list, Total: total, Page: page, Size: size})
}

func (h *Handler) GetActiveLinks(c *gin.Context) {
	list, err := h.svc.Link.ListActive()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, list)
}

func (h *Handler) CreateLink(c *gin.Context) {
	var req dto.AdminLinkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	link, err := h.svc.Link.Create(req.Name, req.Url, req.Description, req.Icon, req.SortOrder)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, link)
}

func (h *Handler) UpdateLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminLinkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	link, err := h.svc.Link.Update(id, req.Name, req.Url, req.Description, req.Icon, req.SortOrder, req.IsActive)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, link)
}

func (h *Handler) DeleteLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Link.Delete(id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteLinks(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Link.BatchDelete(req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}