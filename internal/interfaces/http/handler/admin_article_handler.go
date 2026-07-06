package handler

import (
	"strconv"

	domarticle "sanmoo-server-go/internal/domain/article"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Admin Article Management Handlers ========================

func toArticle(req dto.AdminArticleCreateRequest) domarticle.Article {
	tags := make([]domarticle.TagRef, 0, len(req.TagIDs))
	for _, id := range req.TagIDs {
		tags = append(tags, domarticle.TagRef{ID: id})
	}
	topics := make([]domarticle.TopicRef, 0, len(req.TopicIDs))
	for _, id := range req.TopicIDs {
		topics = append(topics, domarticle.TopicRef{ID: id})
	}
	return domarticle.Article{
		Title:       req.Title,
		TitleImage:  req.TitleImage,
		Description: req.Description,
		Content:     req.Content,
		CategoryID:  req.CategoryID,
		Tags:        tags,
		Topics:      topics,
		IsTop:       req.IsTop == 1,
		IsPublished: req.IsPublished == 1,
	}
}

func (h *Handler) GetArticles(c *gin.Context) {
	var q dto.ArticleListQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.Article.ListArticles(c.Request.Context(), q.Page, q.Size, q.Keyword, q.CategoryID, q.TagID, q.IsPublished)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) CreateArticle(c *gin.Context) {
	var req dto.AdminArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	id, err := h.svc.Article.CreateArticle(c.Request.Context(), toArticle(req))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.IDResponse{ID: id})
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req dto.AdminArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.UpdateArticle(c.Request.Context(), id, toArticle(req)); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) UpdateArticleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	var req struct {
		IsPublished *bool `json:"isPublished"`
		IsTop       *bool `json:"isTop"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.UpdateArticleStatus(c.Request.Context(), id, req.IsPublished, req.IsTop); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchUpdateArticleStatus(c *gin.Context) {
	var req struct {
		IDs         []uint64 `json:"ids" binding:"required"`
		IsPublished *bool    `json:"isPublished"`
		IsTop       *bool    `json:"isTop"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.BatchUpdateArticleStatus(c.Request.Context(), req.IDs, req.IsPublished, req.IsTop); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.DeleteArticle(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) BatchDeleteArticles(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	if err := h.svc.Article.BatchDeleteArticles(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, dto.EmptyResponse{})
}

func (h *Handler) AdminArticleDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Article.GetArticleDetail(c.Request.Context(), id, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}
